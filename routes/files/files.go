/*
 * Nanocloud Community, a comprehensive platform to turn any application
 * into a cloud solution.
 *
 * Copyright (C) 2016 Nanocloud Software
 *
 * This file is part of Nanocloud community.
 *
 * Nanocloud community is free software; you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * Nanocloud community is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/labstack/echo"
	"github.com/manyminds/api2go/jsonapi"
	log "github.com/sirupsen/logrus"
)

type hash map[string]interface{}

type file struct {
	Id      string `json:"-"`
	ModTime int64  `json:"mod_time"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Type    string `json:"type"`
}

func (f *file) GetID() string {
	return f.Id
}

func Patch(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	username := r.URL.Query()["username"][0]
	filename := r.URL.Query()["filename"][0]
	newfilename := r.URL.Query()["newfilename"][0]
	var path string
	var newPath string

	path = fmt.Sprintf("/home/%s/%s", username, filename)
	newPath = fmt.Sprintf("/home/%s/%s", username, newfilename)

	_, err := os.Stat(newPath)
	// if a file with exactly the same name already exists
	if err == nil {
		// we rename it like 'file (2).txt'
		for i := 1; i > 0; i++ {
			extension := filepath.Ext(newPath)
			new_file := newPath[0:len(newPath)-len(extension)] + " (" + strconv.Itoa(i) + ")" + extension
			_, err = os.Stat(new_file)
			if err != nil {
				err = os.Rename(path, new_file)
				if err != nil {
					log.Error(err)
				}
				break
			}
		}
	} else {
		err = os.Rename(path, newPath)
	}

	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to move file", http.StatusInternalServerError)
		return err
	}

	user, err := user.Lookup(username)
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to retrieve user's permission", http.StatusInternalServerError)
		return err
	}

	uid, err := strconv.Atoi(user.Uid)
	gid, err := strconv.Atoi(user.Gid)
	err = os.Chown(newPath, uid, gid)
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to update file permission", http.StatusInternalServerError)
		return err
	}
	return nil
}

func CreateDirectory(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	username := r.URL.Query()["username"][0]
	filename := r.URL.Query()["filename"][0]
	var path string

	path = fmt.Sprintf("/home/%s/%s", username, filename)
	err := os.Mkdir(path, 0777)

	user, err := user.Lookup(username)
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to retrieve user's permission", http.StatusInternalServerError)
		return err
	}

	uid, err := strconv.Atoi(user.Uid)
	gid, err := strconv.Atoi(user.Gid)
	err = os.Chown(path, uid, gid)
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to update file permission", http.StatusInternalServerError)
		return err
	}
	return nil
}

func Delete(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	username := r.URL.Query()["username"][0]
	filename := r.URL.Query()["filename"][0]
	var path string

	path = fmt.Sprintf("/home/%s/%s", username, filename)
	err := os.RemoveAll(path)

	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to remove", http.StatusInternalServerError)
		return err
	}
	return nil
}

func Post(c echo.Context) error {
	var dst *os.File
	r := c.Request()
	w := c.Response()
	username := r.URL.Query()["username"][0]
	filename := r.URL.Query()["filename"][0]
	var path string

	if runtime.GOOS == "windows" {
		dstDir := fmt.Sprintf(`C:\Users\%s\Desktop\Nanocloud`, username)
		err := os.MkdirAll(dstDir, 0777)

		if err != nil {
			log.Error(err)
			http.Error(w, "Unable to create destination directory", http.StatusInternalServerError)
			return err
		}

		path = fmt.Sprintf(`%s\%s`, dstDir, filename)
	} else {
		path = fmt.Sprintf("/home/%s/%s", username, filename)
	}

	_, err := os.Stat(path)
	// if a file with exactly the same name already exists
	if err == nil {
		// we rename it like 'file (2).txt'
		for i := 1; i > 0; i++ {
			extension := filepath.Ext(path)
			new_file := path[0:len(path)-len(extension)] + " (" + strconv.Itoa(i) + ")" + extension
			_, err = os.Stat(new_file)
			if err != nil {
				dst, err = os.Create(new_file)
				if err != nil {
					log.Error(err)
				}
				break
			}
		}
	} else {
		dst, err = os.Create(path)
	}
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to create destination file", http.StatusInternalServerError)
		return err
	}

	defer dst.Close()

	_, err = io.Copy(dst, r.Body)
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to write destination file", http.StatusInternalServerError)
		return err
	}
	dst.Sync()

	user, err := user.Lookup(username)
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to retrieve user's permission", http.StatusInternalServerError)
		return err
	}

	uid, err := strconv.Atoi(user.Uid)
	gid, err := strconv.Atoi(user.Gid)
	err = os.Chown(path, uid, gid)
	if err != nil {
		log.Error(err)
		http.Error(w, "Unable to update file permission", http.StatusInternalServerError)
		return err
	}
	return nil
}

func Get(c echo.Context) error {
	filepath := c.QueryParam("path")
	showHidden := c.QueryParam("show_hidden") == "true"
	create := c.QueryParam("create") == "true"

	if len(filepath) < 1 {
		return c.JSON(
			http.StatusBadRequest,
			hash{
				"error": "Path not specified",
			},
		)
	}

	s, err := os.Stat(filepath)
	if err != nil {
		log.Error(err.(*os.PathError).Err.Error())
		m := err.(*os.PathError).Err.Error()
		if m == "no such file or directory" || m == "The system cannot find the file specified." {
			if create {
				err := os.MkdirAll(filepath, 0777)
				if err != nil {
					return err
				}
				s, err = os.Stat(filepath)
				if err != nil {
					return err
				}
			} else {
				return c.JSON(
					http.StatusNotFound,
					hash{
						"error": "no such file or directory",
					},
				)
			}
		} else {
			return err
		}
	}

	if s.Mode().IsDir() {
		f, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer f.Close()

		files, err := f.Readdir(-1)
		if err != nil {
			return err
		}

		rt := make([]*file, 0)

		for _, fi := range files {
			name := fi.Name()
			if !showHidden && isFileHidden(fi) {
				continue
			}

			fullpath := path.Join(filepath, name)
			id, err := loadFileId(fullpath)
			if err != nil {
				log.Errorf("Cannot retrieve file id for file: %s: %s", fullpath, err.Error())
				continue
			}

			f := &file{
				Id:      id,
				ModTime: fi.ModTime().Unix(),
				Name:    name,
				Size:    fi.Size(),
			}
			if fi.IsDir() {
				f.Type = "directory"
			} else {
				f.Type = "regular file"
			}
			rt = append(rt, f)
		}
		/*
		 * The Content-Length is not set is the buffer length is more than 2048
		 */
		b, err := jsonapi.Marshal(rt)
		if err != nil {
			log.Error(err)
			return err
		}

		r := c.Response()
		r.Header().Set("Content-Length", strconv.Itoa(len(b)))
		r.Header().Set("Content-Type", "application/json; charset=utf-8")
		r.Write(b)
		return nil
	}

	return c.File(filepath + s.Name())
}
