// Copyright (c) Liam Stanley <me@liamstanley.io>. All rights reserved. Use
// of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/dave/jennifer/jen"
)

var reStripOrigin = regexp.MustCompile("[^a-zA-Z0-9_]+")

func gen(path, pkg string, data map[string]*DataNode) error {
	logger.Println("beginning generation from data types now")

	for name := range data {
		f := jen.NewFile(name)
		multiPkgCommentf(f, `
Package %s provides constants and variables based on irc
definitions located here:
  [%s]
  - https://defs.ircdocs.horse/defs/%s.html
  - %s/blob/%s/%s

  Data revision: v%s
`, name, data[name].Data.Page.Name, name, strings.TrimSuffix(flags.Git.Repo, ".git"), flags.Git.Branch, data[name].Path, data[name].Data.Info.Revision)

		ok, err := genParseValues(f, name, data[name])
		if err != nil {
			logger.Printf("error generating pkg %q: %v", name, err)
			delete(data, name)
			continue
		}

		if !ok {
			continue
		}

		dir := filepath.Join(path, name)
		logger.Printf("creating directory %v", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Fatalf("error creating dir %v: %v", dir, err)
		}

		writeFile(f, filepath.Join(dir, name+".go"))
	}

	logger.Printf("generated %d packages", len(data))
	if len(data) == 0 {
		logger.Fatal("at least one package should have been generated")
	}

	logger.Println("reading README_TPL.md")
	readme, err := ioutil.ReadFile(filepath.Join(path, "README_TPL.md"))
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("generating README.md from README_TPL.md")
	tpl := template.Must(template.New("readme").Parse(string(readme)))

	f, err := os.Create(filepath.Join(path, "README.md"))
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()

	if err = tpl.Execute(f, map[string]interface{}{"data": data}); err != nil {
		logger.Fatalf("error generating template from README_TPL.md: %v", err)
	}

	logger.Println("successfully wrote README.md")

	return nil
}

func genParseValues(f *jen.File, pkg string, node *DataNode) (ok bool, err error) {
	switch pkg {
	case "chanmembers":
		if name, ok := node.Data.Format["prefixchar"]; ok {
			f.Comment("ChanPermPrefix is the " + name.(string) + ".")
		}
		f.Type().Id("ChanPermPrefix").Rune()
		if name, ok := node.Data.Format["mode"]; ok {
			f.Comment("ChanPermMode is the " + name.(string) + ".")
		}
		f.Type().Id("ChanPermMode").Rune()

		sort.Slice(node.Data.Values, func(i, j int) bool {
			return node.Data.Values[i].GetString("name") < node.Data.Values[j].GetString("name")
		})

		fields := []jen.Code{}
		for _, val := range node.Data.Values {
			if _, ok := val["name"]; !ok {
				return false, fmt.Errorf("value %#v doesn't have name", val)
			}

			name := strcase.ToCamel(strings.ToLower(val.GetString("name")))
			def := convertMultiComment(splitCommentf(
				0, "%s -- %s\n - obsolete: %t\n - origin: %q",
				name, val.GetString("comment"), val.GetBool("obsolete"),
				val.GetStringFallback("origin", "n/a"),
			))
			def.Id(name).Id("ChanPermPrefix").Op("=").LitRune(val.GetRune("prefixchar"))

			mdef := jen.Commentf("%sMode -- mode mapping to prefix %q.", name, name).Line()
			mdef.Id(name + "Mode").Id("ChanPermMode").Op("=").LitRune(val.GetRune("mode"))
			fields = append(fields, def, mdef)
		}

		f.Const().Defs(fields...)
		return true, nil
	case "chanmodes":
		if name, ok := node.Data.Format["char"]; ok {
			f.Comment("ChanMode is the " + name.(string) + ".")
		}
		f.Type().Id("ChanMode").Rune()

		sort.Slice(node.Data.Values, func(i, j int) bool {
			if node.Data.Values[i].GetString("name") == node.Data.Values[j].GetString("name") {
				if strings.Contains(node.Data.Values[i].GetString("origin"), "RFC") {
					return true
				}
				if node.Data.Values[i].GetBool("conflict") {
					return false
				}
				if node.Data.Values[j].GetBool("conflict") {
					return true
				}
			}
			return node.Data.Values[i].GetString("name") < node.Data.Values[j].GetString("name")
		})

		// TODO: how to address conflicts? want to support something like this.
		// Maybe make a map, and return a list that match?..
		//
		// modeFields := []jen.Code{}
		// for _, val := range node.Data.Values {
		// 	if _, ok := val["name"]; !ok {
		// 		return false, fmt.Errorf("value %#v doesn't have name", val)
		// 	}
		// 	name := strcase.ToCamel(strings.ToLower(val.GetString("name")))
		// 	modeFields = append(modeFields, jen.Case(jen.LitRune(val.GetRune("char"))).Block(
		// 		jen.Return(jen.Id(name), jen.True()),
		// 	))
		// }
		//
		// modeFields = append(modeFields, jen.Default().Block(jen.Return(jen.Id("ChanMode").Values(), jen.False())))
		//
		// f.Comment("IsChanMode is used to see if the mode matches any of the types defined here.")
		// f.Func().Id("IsChanMode").Params(
		// 	jen.Id("mode").Rune(),
		// ).Params(jen.Id("ChanMode"), jen.Bool()).Block(
		// 	jen.Switch(jen.Id("mode")).Block(modeFields...),
		// )

		fields := []jen.Code{}
		for _, val := range node.Data.Values {
			name := strcase.ToCamel(strings.ToLower(val.GetString("name")))
			if val.GetBool("conflict") && !strings.Contains(val.GetString("origin"), "RFC") {
				origin := val.GetString("origin")

				name = strcase.ToCamel(strings.ToLower(val.GetString("name") + "_" + reStripOrigin.ReplaceAllString(origin, "_")))
			}

			comment := fmt.Sprintf(
				"%s (%s) -- %s\n\n - conflict: %t\n - origin: %q",
				name, val.GetString("name"), val.GetString("comment"),
				val.GetBool("conflict"), val.GetStringFallback("origin", "unknown"),
			)
			if value := val.GetString("parameter"); value != "" {
				comment += fmt.Sprintf("\n - parameter: %q", value)
			}

			def := convertMultiComment(splitComment(0, comment))
			def.Id(name).Id("ChanMode").Op("=").LitRune(val.GetRune("char"))

			fields = append(fields, def)
		}

		f.Const().Defs(fields...)
		return true, nil
	case "chantypes":
		if name, ok := node.Data.Format["prefixchar"]; ok {
			f.Comment("ChanType is the " + name.(string) + ".")
		}
		f.Type().Id("ChanType").Rune()

		sort.Slice(node.Data.Values, func(i, j int) bool {
			return node.Data.Values[i].GetString("name") < node.Data.Values[j].GetString("name")
		})

		fields := []jen.Code{}
		for _, val := range node.Data.Values {
			if _, ok := val["name"]; !ok {
				return false, fmt.Errorf("value %#v doesn't have name", val)
			}

			name := strcase.ToCamel(strings.ToLower(val.GetString("name")))
			def := convertMultiComment(splitCommentf(
				0, "%s -- %s\n - origin: %q",
				name, val.GetString("comment"), val.GetStringFallback("origin", "n/a"),
			))
			def.Id(name).Id("ChanType").Op("=").LitRune(val.GetRune("prefixchar"))

			fields = append(fields, def)
		}

		f.Const().Defs(fields...)
		return true, nil
	case "numerics":
		if name, ok := node.Data.Format["char"]; ok {
			f.Comment("Numeric is the " + name.(string) + ".")
		}
		f.Type().Id("Numeric").Int()

		sort.Slice(node.Data.Values, func(i, j int) bool {
			if node.Data.Values[i].GetString("name") == node.Data.Values[j].GetString("name") {
				if strings.Contains(node.Data.Values[i].GetString("origin"), "RFC") {
					return true
				}
				if node.Data.Values[i].GetBool("conflict") {
					return false
				}
				if node.Data.Values[j].GetBool("conflict") {
					return true
				}
			}
			return node.Data.Values[i].GetString("name") < node.Data.Values[j].GetString("name")
		})

		nameCache := make(map[string]int)
		for _, val := range node.Data.Values {
			nameCache[val.GetString("name")]++
		}
		fields := []jen.Code{}
		for _, val := range node.Data.Values {
			name := val.GetString("name")
			if !strings.Contains(val.GetString("origin"), "RFC") && nameCache[name] > 1 {
				origin := val.GetString("origin")

				if _, ok := nameCache[name]; ok {
					name = val.GetString("name") + "_" + reStripOrigin.ReplaceAllString(origin, "_")
				}
			}

			comment := fmt.Sprintf(
				"%s (%s) -- %s\n\n - conflict: %t\n - origin: %q",
				name, val.GetString("name"), val.GetString("comment"),
				val.GetBool("conflict"), val.GetStringFallback("origin", "unknown"),
			)
			if value := val.GetBool("obsolete"); value {
				comment += fmt.Sprintf("\n - obsolete: %t", value)
			}
			if value := val.GetString("information"); value != "" {
				comment += fmt.Sprintf("\n - more info: %s", value)
			}
			if value := val.GetString("seealso"); value != "" {
				comment += fmt.Sprintf("\n - see also: %s", value)
			}
			if value := val.GetString("format"); value != "" {
				comment += fmt.Sprintf("\n - format: %q", value)
			}

			def := convertMultiComment(splitComment(0, comment))
			def.Id(name).Id("Numeric").Op("=").Lit(val.GetInt("numeric"))

			fields = append(fields, def)
		}

		f.Const().Defs(fields...)
		return true, nil
	default:
		return false, errors.New("skipping data type, not known")
	}

}
