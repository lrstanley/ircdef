// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type DataNode struct {
	Hash   plumbing.Hash
	Path   string
	Commit *object.Commit
	Data   *FileContent
}

type FileContent struct {
	Info struct {
		Typ      string `yaml:"type"`
		Revision string `yaml:"revision"`
	} `yaml:"file_info"`
	Page struct {
		Name string `yaml:"name"`
	} `yaml:"page"`
	Format FileValue   `yaml:"format"`
	Values []FileValue `yaml:"values"`
}

type FileValue map[string]interface{}

func (f FileValue) GetString(key string) string {
	if val, ok := f[key].(string); ok {
		return val
	}
	return ""
}
func (f FileValue) GetStringFallback(key, fallback string) string {
	if out := f.GetString(key); out != "" {
		return out
	}
	return fallback
}
func (f FileValue) GetRune(key string) rune {
	if val, ok := f[key].(rune); ok {
		return val
	}
	if val, ok := f[key].(string); ok {
		return rune(val[0])
	}
	return 0
}
func (f FileValue) GetBool(key string) bool {
	if val, ok := f[key].(bool); ok {
		return val
	}
	return false
}
func (f FileValue) GetInt(key string) int {
	if val, ok := f[key].(int); ok {
		return val
	}
	if val, ok := f[key].(string); ok {
		if val, err := strconv.Atoi(val); err == nil {
			return val
		}
	}
	return -1
}

func fetchData(uri, branch string) (map[string]*DataNode, error) {
	logger.Printf("fetching %s (branch: %s)", uri, branch)
	stor := memory.NewStorage()
	repo, err := git.Clone(stor, nil, &git.CloneOptions{
		URL:               uri,
		ReferenceName:     plumbing.ReferenceName("refs/heads/" + branch),
		SingleBranch:      true,
		Depth:             1,
		RecurseSubmodules: git.NoRecurseSubmodules,
		Progress:          os.Stdout,
		NoCheckout:        true,
		Tags:              git.NoTags,
	})
	if err != nil {
		return nil, err
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}
	logger.Printf("using ref: %s", ref.Hash().String())
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	logger.Println("iterating through object tree")
	seen := make(map[plumbing.Hash]bool)
	iter := object.NewTreeWalker(tree, true, seen)

	data := map[string]*DataNode{}

	var file *object.File
	var ferr error
	var name string
	var entry object.TreeEntry

	for err == nil {
		name, entry, err = iter.Next()
		if strings.HasPrefix(name, "_data/") && !strings.HasPrefix(name, "_data/validation") && strings.HasSuffix(name, ".yaml") {
			logger.Printf("parsing %v", name)

			file, ferr = tree.File(name)
			if ferr != nil {
				logger.Printf("error reading object %s (path: %v): %v", entry.Hash.String(), name, ferr)
				continue
			}

			r, ferr := file.Reader()
			if err != nil {
				logger.Printf("error reading object %s (path: %v): %v", entry.Hash.String(), name, ferr)
			}

			fc := &FileContent{}

			dec := yaml.NewDecoder(r)
			if ferr = dec.Decode(fc); ferr != nil {
				logger.Printf("error unmarshalling yaml at %v: %v", name, ferr)
			}

			node := &DataNode{
				Hash:   entry.Hash,
				Path:   name,
				Commit: commit,
				Data:   fc,
			}

			logger.Printf("parsed %v (%dkb); adding to data array", name, file.Size/1024)
			_, fname := filepath.Split(name)
			data[strings.TrimSuffix(fname, ".yaml")] = node
		}
	}

	logger.Println("finished iterating through object tree")
	if err != io.EOF {
		return nil, err
	}

	return data, nil
}
