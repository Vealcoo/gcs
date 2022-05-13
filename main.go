package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"
)

const (
	projectID  = "projectID"
	bucketName = "bucketName"
)

type Client struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	Path       string
}

var client *Client

func init() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
	conn, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	client = &Client{
		cl:         conn,
		bucketName: bucketName,
		projectID:  projectID,
		Path:       "test-files/",
	}
}
func main() {
	g := gin.Default()
	g.POST("/upload", upload)
	g.POST("/delete", delete)
	g.POST("/list", list)

	g.Run(":8887")
}

// routerAction
func upload(g *gin.Context) {
	f, err := g.FormFile("file_input")
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	tempFile, err := f.Open()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = client.UploadFile(tempFile, f.Filename)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	g.JSON(200, gin.H{
		"message": "success",
	})
}

type DeleteInfo struct {
	Filename string
}

func delete(g *gin.Context) {
	info := DeleteInfo{}
	g.BindJSON(&info)
	err := client.DeleteFile(info.Filename)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	g.JSON(200, gin.H{
		"message": "success",
	})
}
func list(g *gin.Context) {
	result, err := client.ListFile()
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	g.JSON(200, gin.H{
		"message": result,
	})
}

// controller
func (c *Client) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	wc := c.cl.Bucket(c.bucketName).Object(c.Path + object).NewWriter(ctx)
	fmt.Println(c.Path + object)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return nil
}
func (c *Client) DeleteFile(object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	err := c.cl.Bucket(c.bucketName).Object(c.Path + object).Delete(ctx)
	fmt.Println(c.Path + object)
	if err != nil {
		return fmt.Errorf("Delete: %v", err)
	}
	return nil
}
func (c *Client) ListFile() ([]string, error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	result := []string{}
	it := c.cl.Bucket(c.bucketName).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return result ,fmt.Errorf("Bucket(%q).Objects: %v", c.bucketName, err)
		}
		result = append(result, attrs.Name)
	}
	return result ,nil
}

