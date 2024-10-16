package darkibox

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "github.com/rclone/rclone/fs"
    "github.com/rclone/rclone/fs/hash"
    "io"
    "os"
)

// Fs représente le backend darkibox
type Fs struct {
    name string
    root string
    apiKey string
}

// NewFs initialise un nouveau backend darkibox
func NewFs(name, root, apiKey string) (fs.Fs, error) {
    f := &Fs{
        name: name,
        root: root,
        apiKey: apiKey,
    }
    return f, nil
}

// Features retourne les fonctionnalités du backend
func (f *Fs) Features() *fs.Features {
    return &fs.Features{}
}

// Hashes retourne les types de hash supportés
func (f *Fs) Hashes() hash.Set {
    return hash.Set(hash.None)
}

// AccountInfo retourne les informations de compte depuis l'API
func (f *Fs) AccountInfo() (map[string]interface{}, error) {
    apiUrl := fmt.Sprintf("https://darkibox.com/api/account/info?key=%s", f.apiKey)
    resp, err := http.Get(apiUrl)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

// GetUploadServer récupère l'URL du serveur d'upload
func (f *Fs) GetUploadServer() (string, error) {
    apiUrl := fmt.Sprintf("https://darkibox.com/api/upload/server?key=%s", f.apiKey)
    resp, err := http.Get(apiUrl)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }
    if serverUrl, ok := result["result"].(string); ok {
        return serverUrl, nil
    }
    return "", fmt.Errorf("server URL not found in response")
}

// UploadFile permet d'uploader un fichier vidéo
func (f *Fs) UploadFile(filePath string, title string, description string) (map[string]interface{}, error) {
    serverUrl, err := f.GetUploadServer()
    if err != nil {
        return nil, err
    }

    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    body := &url.Values{}
    body.Set("key", f.apiKey)
    body.Set("file_title", title)
    body.Set("file_descr", description)

    request, err := http.NewRequest("POST", serverUrl, io.MultiReader(file))
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(request)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

// FileInfo récupère les informations sur un fichier donné
func (f *Fs) FileInfo(fileCode string) (map[string]interface{}, error) {
    apiUrl := fmt.Sprintf("https://darkibox.com/api/file/info?key=%s&file_code=%s", f.apiKey, fileCode)
    resp, err := http.Get(apiUrl)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}
