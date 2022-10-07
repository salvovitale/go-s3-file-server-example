package web

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/rs/zerolog/log"
	"github.com/salvovitale/go-s3-file-server-example/internal/store"
)

type s3FileHandler interface {
	UploadFile(bucketName, objectName string, file io.Reader, size int64) error
	RemoveFile(bucketName, objectName string) error
	DownloadFile(bucketName, objectName string) (io.Reader, error)
}

type dbFileHandler interface {
	StoreFile(file *store.File) error
	DeleteFile(id uuid.UUID) error
	File(id uuid.UUID) (store.File, error)
}
type FileHandler struct {
	dbHandler  dbFileHandler
	s3Handler  s3FileHandler
	bucketName string
}

func (h *FileHandler) uploadView() http.HandlerFunc {
	type data struct {
		// SessionData
		CSRF template.HTML // string which is not escaped
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/upload.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, data{
			// SessionData: GetSessionData(h.sessions, r.Context()),
			CSRF: csrf.TemplateField(r),
		})
	}
}

func (h *FileHandler) upload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("upload file endpoint called")

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		r.ParseMultipartForm(10 << 20)
		description := r.FormValue("description")
		file, handler, err := r.FormFile("myFile")
		if err != nil {
			log.Err(err).Msg("Error Retrieving the File")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		log.Info().Str("file-name", handler.Filename).Msg("filename")
		log.Info().Int64("file-size", handler.Size).Msg("filesize")
		log.Info().Str("file-header", fmt.Sprintf("%v", handler.Header)).Msg("MIME Header")
		log.Info().Str("description", description).Msg("description")

		fileUUID := uuid.New()
		err = h.dbHandler.StoreFile(&store.File{ID: fileUUID, FileName: handler.Filename, Description: description})
		if err != nil {
			log.Err(err).Msg("Error storing file in db")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info().Str("file-uuid", fileUUID.String()).Str("filename", handler.Filename).Msg("Stored file in db")

		err = h.s3Handler.UploadFile(h.bucketName, fileUUID.String(), file, handler.Size)
		if err != nil {
			log.Err(err).Msg("Error uploading file to S3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info().Msg("Successfully Uploaded File to S3")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (h *FileHandler) delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//parse the id
		idStr := chi.URLParam(r, "id")

		//parse and validate the id
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Error().Err(err).Msg("Error parsing id")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//delete file from db
		if err := h.dbHandler.DeleteFile(id); err != nil {
			log.Error().Err(err).Msg("Error deleting file from db")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// delete file from s3
		if err := h.s3Handler.RemoveFile(h.bucketName, id.String()); err != nil {
			log.Error().Err(err).Msg("Error deleting file from s3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info().Msgf("Successfully deleted file %s from s3", idStr)

		// redirect to the thread list
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (h *FileHandler) download() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//parse the id
		idStr := chi.URLParam(r, "id")

		//parse and validate the id
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Error().Err(err).Msg("Error parsing id")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		//retrieve file from db
		f, err := h.dbHandler.File(id)
		if err != nil {
			log.Error().Err(err).Msg("Error retrieving file from db")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// delete file from s3
		fio, err := h.s3Handler.DownloadFile(h.bucketName, id.String())
		if err != nil {
			log.Error().Err(err).Msg("Error downloading file from s3")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// read file into byte array
		fileBytes, err := ioutil.ReadAll(fio)
		if err != nil {
			log.Error().Err(err).Msg("Error converting file from s3 into byte array")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.FileName))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Write(fileBytes)
		// redirect to the thread list
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// This method shows how to add a file locally to the server. It is not used in the example. It is here for reference.
func (h *FileHandler) uploadFileToServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("upload file endpoint called")

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		r.ParseMultipartForm(10 << 20)

		file, handler, err := r.FormFile("myFile")
		if err != nil {
			log.Err(err).Msg("Error Retrieving the File")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		log.Info().Str("file-name", handler.Filename).Msg("filename")
		log.Info().Int64("file-size", handler.Size).Msg("filesize")
		log.Info().Str("file-header", fmt.Sprintf("%v", handler.Header)).Msg("MIME Header")

		// Create a temporary file within our temp-images directory that follows
		// a particular naming pattern
		tempFile, err := ioutil.TempFile("files_uploaded", "upload-*.docx")
		if err != nil {
			log.Err(err).Msg("Error Creating temporary file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tempFile.Close()
		// read all of the contents of our uploaded file into a
		// byte array
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Err(err).Msg("Error creating byte array from uploaded file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// write this byte array to our temporary file
		_, err = tempFile.Write(fileBytes)
		if err != nil {
			log.Err(err).Msg("Error copying uploaded file into temporary file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Info().Msg("Successfully Uploaded File to server")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
