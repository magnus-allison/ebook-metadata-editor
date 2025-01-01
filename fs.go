package main

import (
	"archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func loadPathName () string {
	file, err := os.Open("path.txt")
  	if err != nil { panic(err) }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var text string
	for scanner.Scan() { text += scanner.Text() }
	return text
}

func openReader (path string) *zip.ReadCloser {
	reader, err := zip.OpenReader(path)
	if err != nil {
		panic("failed to open EPUB: " + err.Error())
	}
	return reader
}

type EPub struct {
	Name      string
	Path      string
	PathNoExt string
	MimeType  string
	FileSize  int64
	OPFData  *OPFData
	CoverImagePath string
}

// contributor, coverage, creator, date, description, format, identifier, language, publisher, relation, rights, source, subject, title, type
type OPFData struct {
	Metadata   struct {
		Title      string `xml:"http://purl.org/dc/elements/1.1/ title"`
		Creator    string `xml:"http://purl.org/dc/elements/1.1/ creator"`
		Language   string `xml:"http://purl.org/dc/elements/1.1/ language"`
		Identifier string `xml:"http://purl.org/dc/elements/1.1/ identifier"`
		Date 	   string `xml:"http://purl.org/dc/elements/1.1/ date"`
	} `xml:"metadata"`
}

type ManifestItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type Package struct {
	Manifest struct {
		Items []ManifestItem `xml:"item"`
	} `xml:"manifest"`
	Metadata struct {
		Meta []struct {
			Name string `xml:"name,attr"`
			Content string `xml:"content,attr"`
		} `xml:"meta"`
	} `xml:"metadata"`
}

// ContainerXML represents the structure of META-INF/container.xml
type ContainerXML struct {
	RootFiles struct {
		RootFile struct {
			FullPath string `xml:"full-path,attr"`
		} `xml:"rootfile"`
	} `xml:"rootfiles"`
}


func CollectBooks (dir string) ([]*EPub, error) {
	files, err := os.ReadDir(dir)
	if err != nil { return nil, err }

	books := []*EPub{}

	for _, file := range files {

		if filepath.Ext(file.Name()) != ".epub" { continue }

		book := &EPub{
			Name: file.Name(),
			Path: filepath.Join(dir, file.Name()),
			PathNoExt: filepath.Join(dir, file.Name()[:len(file.Name())-5]),
			MimeType: "unknown",
		}
		book.SetOPFData()
		book.SetCoverImagePath()
		books = append(books, book)
	}

	// sort by title
	sort.Slice(books, func(i, j int) bool {
		return books[i].OPFData.Metadata.Title < books[j].OPFData.Metadata.Title
	})

	return books, nil
}


func (e *EPub) SetOPFData() {

	reader := openReader(e.Path)
	defer reader.Close()

	opfPath, err := getOpfPath(reader)
	if err != nil {
		fmt.Errorf("failed to locate content.opf: %v", err)
	}

	var metadata *OPFData
	for _, file := range reader.File {
		if file.Name == "mimetype" {
			rc, _ := file.Open()
			defer rc.Close()
			scanner := bufio.NewScanner(rc)
			scanner.Scan()
			e.MimeType = scanner.Text()
		}
		// fmt.Println(file.Name)
		if file.Name == opfPath {
			rc, err := file.Open()
			if err != nil {
				fmt.Errorf("failed to open content.opf: %v", err)
			}
			defer rc.Close()

			var opf OPFData
			if err := xml.NewDecoder(rc).Decode(&opf); err != nil {
				fmt.Errorf("failed to parse content.opf: %v", err)
			}

			fmt.Println(opf)

			// Populate the MetaData struct.
			metadata = &opf
			break
		}
	}

	if metadata == nil {
		fmt.Errorf("metadata not found in EPUB")
	}


	if metadata.Metadata.Title == "" {
		metadata.Metadata.Title = "UNKNOWN"
	}

	e.OPFData = metadata
}

func (e *EPub) SetCoverImagePath() {

	reader := openReader(e.Path)
	defer reader.Close()

	// Locate the content.opf file.
	opfPath, err := getOpfPath(reader)
	if err != nil {
		fmt.Errorf("failed to locate content.opf: %v", err)
	}

	// Read and parse the content.opf file.
	var coverID, coverHref string
	for _, file := range reader.File {
		if file.Name == opfPath {
			rc, err := file.Open()
			if err != nil {
				fmt.Errorf("failed to open content.opf: %v", err)
			}
			defer rc.Close()

			var pkg Package
			if err := xml.NewDecoder(rc).Decode(&pkg); err != nil {
				fmt.Errorf("failed to parse content.opf: %v", err)
			}

			// Find the cover ID in metadata.


			// for _, smth := range pkg.Metadata {
			// 	fmt.Println(smth)
			// }

			for _, meta := range pkg.Metadata.Meta {

				// fmt.Println("=====================================")
				// fmt.Println(meta)

				if meta.Name == "cover" {
					coverID = meta.Content
					break
				}
			}

			// Find the cover file path in the manifest.
			for _, item := range pkg.Manifest.Items {
				if item.ID == coverID && (item.MediaType == "image/jpeg" || item.MediaType == "image/png") {
					coverHref = item.Href
					break
				}
			}
			break
		}
	}

	if coverHref == "" {
		fmt.Errorf("cover image not found in EPUB")
	}

	// Absolute path
	pathAbs := filepath.Join(e.PathNoExt, coverHref)
	e.CoverImagePath = pathAbs

}

func (e *EPub) UpdateCoverImage(imagePath string) error {

	reader := openReader(e.Path)
	defer reader.Close()

	// Create a temporary output file for the updated EPUe.
	tempFilePath := e.Path + ".tmp"
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	// Create a new ZIP writer.
	writer := zip.NewWriter(tempFile)

	// Locate content.opf and copy existing files to the new archive.
	var opfPath string
	var newImageName = filepath.Base(imagePath)
	contentRewritten := false

	for _, file := range reader.File {
		// Check if this is the content.opf file.
		if file.Name == "META-INF/container.xml" {
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open container.xml: %v", err)
			}
			defer rc.Close()

			var container ContainerXML
			if err := xml.NewDecoder(rc).Decode(&container); err != nil {
				return fmt.Errorf("failed to parse container.xml: %v", err)
			}

			opfPath = container.RootFiles.RootFile.FullPath
		}

		// Copy all files except the cover image (if it exists).
		if file.Name != opfPath {
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file: %v", err)
			}
			defer rc.Close()

			wc, err := writer.Create(file.Name)
			if err != nil {
				return fmt.Errorf("failed to create file in new archive: %v", err)
			}

			if _, err := io.Copy(wc, rc); err != nil {
				return fmt.Errorf("failed to copy file: %v", err)
			}
		}

		// Handle content.opf file modification.
		if file.Name == opfPath {
			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open content.opf: %v", err)
			}
			defer rc.Close()

			var pkg Package
			if err := xml.NewDecoder(rc).Decode(&pkg); err != nil {
				return fmt.Errorf("failed to parse content.opf: %v", err)
			}

			// Update the metadata and manifest for the new cover image.
			coverUpdated := false
			for i, meta := range pkg.Metadata.Meta {
				if meta.Name == "cover" {
					pkg.Metadata.Meta[i].Content = "cover-image"
					coverUpdated = true
					break
				}
			}
			if !coverUpdated {
				pkg.Metadata.Meta = append(pkg.Metadata.Meta, struct {
					Name    string `xml:"name,attr"`
					Content string `xml:"content,attr"`
				}{
					Name:    "cover",
					Content: "cover-image",
				})
			}

			manifestUpdated := false
			for i, item := range pkg.Manifest.Items {
				if item.ID == "cover-image" {
					pkg.Manifest.Items[i].Href = newImageName
					pkg.Manifest.Items[i].MediaType = "image/jpeg"
					manifestUpdated = true
					break
				}
			}
			if !manifestUpdated {
				pkg.Manifest.Items = append(pkg.Manifest.Items, ManifestItem{
					ID:        "cover-image",
					Href:      newImageName,
					MediaType: "image/jpeg",
				})
			}

			// Write the updated content.opf to the new archive.
			wc, err := writer.Create(file.Name)
			if err != nil {
				return fmt.Errorf("failed to create updated content.opf: %v", err)
			}

			encoder := xml.NewEncoder(wc)
			encoder.Indent("", "  ")
			if err := encoder.Encode(pkg); err != nil {
				return fmt.Errorf("failed to encode updated content.opf: %v", err)
			}
			contentRewritten = true
		}
	}

	// Add the new cover image.
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open new cover image: %v", err)
	}
	defer imageFile.Close()

	wc, err := writer.Create(newImageName)
	if err != nil {
		return fmt.Errorf("failed to add new cover image to archive: %v", err)
	}
	if _, err := io.Copy(wc, imageFile); err != nil {
		return fmt.Errorf("failed to copy new cover image: %v", err)
	}

	// Close the writer.
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close new EPUB archive: %v", err)
	}

	// Replace the original EPUB with the updated file.
	if err := os.Rename(tempFilePath, e.Path); err != nil {
		return fmt.Errorf("failed to replace original EPUB: %v", err)
	}

	if !contentRewritten {
		return fmt.Errorf("content.opf was not updated")
	}

	return nil
}


// Read META-INF/container.xml to locate content.opf
func getOpfPath (reader *zip.ReadCloser) (string, error) {
	var opfPath string
	for _, file := range reader.File {
		if file.Name == "META-INF/container.xml" {
			rc, err := file.Open()
			if err != nil {
				panic("failed to open container.xml: " + err.Error())
			}
			defer rc.Close()

			// Parse cont  ainer.xml
			var container ContainerXML
			if err := xml.NewDecoder(rc).Decode(&container); err != nil {
				panic("failed to parse container.xml: " + err.Error())
			}

			opfPath = container.RootFiles.RootFile.FullPath
			break
		}
	}

	if opfPath == "" {
		return "", fmt.Errorf("content.opf not found in EPUB")
	}

	return opfPath, nil
}