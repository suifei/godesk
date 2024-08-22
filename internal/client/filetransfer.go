package client

import (
    "io"
    "os"
    "path/filepath"
)

func SendFile(conn io.Writer, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(conn, file)
    return err
}

func ReceiveFile(conn io.Reader, filePath string) error {
    os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
    
    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(file, conn)
    return err
}