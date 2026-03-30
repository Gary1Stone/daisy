package util

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// Image resizer for jpeg including Exif (orientation)
func InitialPhotoUploadResizer(filename string) {
	//Get directory paths to the files
	picWidth := 1024
	serverDir := getServerDir()
	photo := filepath.Join(serverDir, "web", "public", "images", filename)
	// Open the image file
	img, err := imaging.Open(photo, imaging.AutoOrientation(true))
	if err != nil {
		log.Println(err)
		return
	}
	// Calculate the aspect ratio
	origWidth := img.Bounds().Dx()
	origHeight := img.Bounds().Dy()
	aspectRatio := float64(origWidth) / float64(origHeight)
	picHeight := int(float64(picWidth) / aspectRatio)
	pic := imaging.Resize(img, picWidth, picHeight, imaging.Lanczos)
	err = imaging.Save(pic, photo, imaging.JPEGQuality(95))
	if err != nil {
		log.Println(err)
	}
}

// Image resizer for jpeg including Exif (orientation)
// Save the thumbnail to a new file ending in -sm.jpg
func resizePhoto(filename string) {
	//Get directory paths to the files
	picWidth := 1024
	thumbnailWidth := 200
	serverDir := getServerDir()
	photo := filepath.Join(serverDir, "web", "public", "images", filename)
	newName := AddSuffixBeforeExtension(filename, "-sm")
	photoSmall := filepath.Join(serverDir, "web", "public", "images", newName)
	// Open the image file
	img, err := imaging.Open(photo, imaging.AutoOrientation(true))
	if err != nil {
		log.Println(err)
		return
	}
	// Calculate the aspect ratio
	origWidth := img.Bounds().Dx()
	origHeight := img.Bounds().Dy()
	aspectRatio := float64(origWidth) / float64(origHeight)
	picHeight := int(float64(picWidth) / aspectRatio)
	thumbnailHeight := int(float64(thumbnailWidth) / aspectRatio)
	// Resize original photo, preserve Exif data when saving
	pic := imaging.Resize(img, picWidth, picHeight, imaging.Lanczos)
	err = imaging.Save(pic, photo, imaging.JPEGQuality(95))
	if err != nil {
		log.Println(err)
	}
	// Create thumbnail of photo
	thumbnail := imaging.Thumbnail(img, thumbnailWidth, thumbnailHeight, imaging.Lanczos)
	err = imaging.Save(thumbnail, photoSmall, imaging.JPEGQuality(95))
	if err != nil {
		log.Println(err)
	}
}

// Get the executable file path to the web server (THIS PROGRAM)
func getServerDir() string {
	executable, err := os.Executable()
	if err != nil {
		log.Println(err)
		return ""
	}
	return filepath.Dir(executable)
}

// Thumbnail name:- Convert string xxxx.jpg to xxxx-sm.jpg
func AddSuffixBeforeExtension(fileName, suffix string) string {
	ext := filepath.Ext(fileName)
	base := fileName[:len(fileName)-len(ext)]
	return base + suffix + ext
}

// Convert all photos in the images directory to 1024x760 resolution
func ShrinkPhotos() {
	dirPath := filepath.Join(getServerDir(), "web", "public", "images")
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println(err)
		return
	}
	// Check if the file is a regular file and ends with ".jpg" and not "-sm.jpg"
	for _, file := range files {
		if file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".jpg") && !strings.HasSuffix(file.Name(), "-sm.jpg") {
			fileInfo, err := file.Info()
			if err != nil {
				log.Println(err)
			}
			// Check file size is greater than 1 Mbyte or xxxxx-sm.jpg does not exist
			if fileInfo.Size() > 1048576 || !isSmallPhotoExisting(files, file.Name()) {
				resizePhoto(file.Name())
			}
		}
	}
}

// Look through directory listing for the small version of the photo
func isSmallPhotoExisting(fnames []fs.DirEntry, target string) bool {
	target = AddSuffixBeforeExtension(target, "-sm")
	for _, fname := range fnames {
		if fname.Name() == target {
			return true
		}
	}
	return false
}

// Return a MAP of all the file names in the images directory
func MapPhotos() map[string]bool {
	imageMap := make(map[string]bool)
	dirPath := filepath.Join(getServerDir(), "web", "public", "images")
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println("ERROR listing files ", err)
		return imageMap
	}
	for _, file := range files {
		if !file.IsDir() {
			imageMap[file.Name()] = true
		}
	}
	return imageMap
}

// Generate a link to a small thumbnail image of the device
func GetThumbnail(img string) string {
	var ctrl strings.Builder
	ctrl.WriteString("<img src='images/")
	if len(img) > 0 {
		ctrl.WriteString(img)
	} else {
		ctrl.WriteString("missing.jpg")
	}
	ctrl.WriteString("' width='100' height='auto'>")
	return ctrl.String()
}

func DeletePhoto(filename string) {
	serverDir := getServerDir()
	photo := filepath.Join(serverDir, "web", "public", "images", filename)
	err := os.Remove(photo)
	if err != nil {
		log.Println("Error deleting file:", err)
	}
}
