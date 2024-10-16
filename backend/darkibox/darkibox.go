package darkibox

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/fserrors"
	"github.com/rclone/rclone/fs/hash"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rclone/rclone/fs/rest"
	"github.com/rclone/rclone/lib/pacer"
	"net/http"
	"path"
)

// Fs represents a remote darkibox filesystem
type Fs struct {
	name     string
	root     string
	features *fs.Features
	srv      *rest.Client
	pacer    *pacer.Pacer
}

// Object represents a remote darkibox file object
type Object struct {
	fs     *Fs
	remote string
	info   *FileInfo
}

// FileInfo holds file metadata
type FileInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	ModTime   string `json:"mod_time"`
	MimeType  string `json:"mime_type"`
	IsDir     bool   `json:"is_dir"`
}

// NewFs creates a new Fs object
func NewFs(name, root string, m configmap.Mapper) (fs.Fs, error) {
	srv := rest.NewClient(fs.Config)
	f := &Fs{
		name: name,
		root: root,
		srv:  srv,
		pacer: pacer.New().SetMinSleep(fs.Config.MinSleep).SetPacer(fs.Config.Pacer),
	}
	f.features = (&fs.Features{
		CanHaveEmptyDirectories: true,
	}).Fill(f)
	return f, nil
}

// List the objects and directories in dir into entries
func (f *Fs) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {
	url := fmt.Sprintf("/list/%s", path.Join(f.root, dir))
	resp, err := f.srv.CallJSON(ctx, &rest.Opts{
		Method: "GET",
		Path:   url,
	}, nil, &entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// NewObject creates a new object
func (f *Fs) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	info := &FileInfo{}
	url := fmt.Sprintf("/fileinfo/%s", path.Join(f.root, remote))
	err := f.srv.CallJSON(ctx, &rest.Opts{
		Method: "GET",
		Path:   url,
	}, nil, info)
	if err != nil {
		return nil, err
	}
	return &Object{fs: f, remote: remote, info: info}, nil
}

// Put uploads a file to the remote
func (f *Fs) Put(ctx context.Context, in fs.ObjectInfo, src fs.Object) (fs.Object, error) {
	url := fmt.Sprintf("/upload/%s", path.Join(f.root, in.Remote()))
	opts := rest.Opts{
		Method:        "POST",
		Path:          url,
		Body:          in,
		ContentLength: in.Size(),
	}
	resp, err := f.srv.Call(ctx, &opts)
	if err != nil {
		return nil, err
	}
	defer fs.CheckClose(resp.Body, &err)
	return f.NewObject(ctx, in.Remote())
}

// Hashes returns the supported hash types
func (f *Fs) Hashes() hash.Set {
	return hash.Set(hash.None)
}

// Features returns the optional features of this Fs
func (f *Fs) Features() *fs.Features {
	return f.features
}

// Mkdir creates a directory
func (f *Fs) Mkdir(ctx context.Context, dir string) error {
	url := fmt.Sprintf("/mkdir/%s", path.Join(f.root, dir))
	opts := rest.Opts{
		Method: "POST",
		Path:   url,
	}
	_, err := f.srv.Call(ctx, &opts)
	return err
}

// Rmdir removes a directory
func (f *Fs) Rmdir(ctx context.Context, dir string) error {
	url := fmt.Sprintf("/rmdir/%s", path.Join(f.root, dir))
	opts := rest.Opts{
		Method: "DELETE",
		Path:   url,
	}
	_, err := f.srv.Call(ctx, &opts)
	return err
}

// Remove deletes a file
func (f *Fs) Remove(ctx context.Context, remote string) error {
	url := fmt.Sprintf("/remove/%s", path.Join(f.root, remote))
	opts := rest.Opts{
		Method: "DELETE",
		Path:   url,
	}
	_, err := f.srv.Call(ctx, &opts)
	return err
}

// Object methods

// Update updates the Object with the contents of the io.Reader
func (o *Object) Update(ctx context.Context, in fs.ObjectInfo, src fs.Object) error {
	url := fmt.Sprintf("/update/%s", path.Join(o.fs.root, o.remote))
	opts := rest.Opts{
		Method:        "PUT",
		Path:          url,
		Body:          in,
		ContentLength: in.Size(),
	}
	resp, err := o.fs.srv.Call(ctx, &opts)
	if err != nil {
		return err
	}
	defer fs.CheckClose(resp.Body, &err)
	return nil
}

// Remove deletes the object
func (o *Object) Remove(ctx context.Context) error {
	return o.fs.Remove(ctx, o.remote)
}

func (o *Object) Size() int64 {
	return o.info.Size
}

func (o *Object) ModTime(ctx context.Context) (modTime fs.Time, err error) {
	return o.info.ModTime, nil
}

func (o *Object) Fs() fs.Fs {
	return o.fs
}

func (o *Object) String() string {
	return o.remote
}
