/*
 * Minio Client (C) 2015 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"

	"github.com/minio/cli"
	"github.com/minio/mc/pkg/client"
)

func checkCopySyntax(ctx *cli.Context) {
	if len(ctx.Args()) < 2 {
		cli.ShowCommandHelpAndExit(ctx, "cp", 1) // last argument is exit code.
	}

	// extract URLs.
	URLs, err := args2URLs(ctx.Args())
	fatalIf(err.Trace(ctx.Args()...), fmt.Sprintf("Argument parsing failed."))

	if len(URLs) < 2 {
		fatalIf(err.Trace(ctx.Args()...), fmt.Sprintf("Unable to parse source and target arguments."))
	}

	srcURLs := URLs[:len(URLs)-1]
	tgtURL := URLs[len(URLs)-1]
	isRecursive := ctx.Bool("recursive")

	/****** Generic rules *******/
	url := client.NewURL(tgtURL)
	if url.Host != "" {
		// This check is for type URL.
		if url.Path == string(url.Separator) {
			fatalIf(errInvalidArgument().Trace(), fmt.Sprintf("Target ‘%s’ does not contain bucket name.", tgtURL))
		}
	}

	switch guessCopyURLType(srcURLs, tgtURL, isRecursive) {
	case copyURLsTypeA: // File -> File.
		checkCopySyntaxTypeA(srcURLs, tgtURL)
	case copyURLsTypeB: // File -> Folder.
		checkCopySyntaxTypeB(srcURLs, tgtURL)
	case copyURLsTypeC: // Folder... -> Folder.
		checkCopySyntaxTypeC(srcURLs, tgtURL)
	case copyURLsTypeD: // File | Folder... -> Folder.
		checkCopySyntaxTypeD(srcURLs, tgtURL)
	default:
		fatalIf(errInvalidArgument().Trace(), "Invalid arguments.")
	}
}

// checkCopySyntaxTypeA verifies if the source and target are valid file arguments.
func checkCopySyntaxTypeA(srcURLs []string, tgtURL string) {
	// Check source.
	if len(srcURLs) != 1 {
		fatalIf(errInvalidArgument().Trace(), "Invalid number of source arguments.")
	}
	srcURL := srcURLs[0]
	_, srcContent, err := url2Stat(srcURL)
	fatalIf(err.Trace(srcURL), "Unable to stat source ‘"+srcURL+"’.")

	if !srcContent.Type.IsRegular() {
		fatalIf(errInvalidArgument().Trace(), "Source ‘"+srcURL+"’ is not a file.")
	}
}

// checkCopySyntaxTypeB verifies if the source is a valid file and target is a valid dir.
func checkCopySyntaxTypeB(srcURLs []string, tgtURL string) {
	// Check source.
	if len(srcURLs) != 1 {
		fatalIf(errInvalidArgument().Trace(), "Invalid number of source arguments.")
	}
	srcURL := srcURLs[0]
	_, srcContent, err := url2Stat(srcURL)
	fatalIf(err.Trace(srcURL), "Unable to stat source ‘"+srcURL+"’.")

	if !srcContent.Type.IsRegular() {
		fatalIf(errInvalidArgument().Trace(srcURL), "Source ‘"+srcURL+"’ is not a file.")
	}

	// Check target.
	if _, tgtContent, err := url2Stat(tgtURL); err == nil {
		if !tgtContent.Type.IsDir() {
			fatalIf(errInvalidArgument().Trace(tgtURL), "Target ‘"+tgtURL+"’ is not a folder.")
		}
	}
}

// checkCopySyntaxTypeC verifies if the source is a valid recursive dir and target is a valid dir.
func checkCopySyntaxTypeC(srcURLs []string, tgtURL string) {
	// Check source.
	if len(srcURLs) != 1 {
		fatalIf(errInvalidArgument().Trace(), "Invalid number of source arguments.")
	}

	srcURL := srcURLs[0]
	_, _, err := url2Stat(srcURL)
	if err != nil && !isURLPrefixExists(srcURL) {
		fatalIf(err.Trace(srcURL), "Unable to stat source ‘"+srcURL+"’.")
	}

	// Check target.
	if _, tgtContent, err := url2Stat(tgtURL); err == nil {
		if !tgtContent.Type.IsDir() {
			fatalIf(errInvalidArgument().Trace(tgtURL), "Target ‘"+tgtURL+"’ is not a folder.")
		}
	}
}

// checkCopySyntaxTypeD verifies if the source is a valid list of files and target is a valid dir.
func checkCopySyntaxTypeD(srcURLs []string, tgtURL string) {
	// Check source.
	for _, srcURL := range srcURLs {
		_, srcContent, err := url2Stat(srcURL)
		fatalIf(err.Trace(srcURL), "Unable to stat source ‘"+srcURL+"’.")

		if !srcContent.Type.IsRegular() {
			fatalIf(errInvalidArgument().Trace(), "Source ‘"+srcURL+"’ is not a file.")
		}
	}

	// Check target.
	if _, tgtContent, err := url2Stat(tgtURL); err == nil {
		if !tgtContent.Type.IsDir() {
			fatalIf(errInvalidArgument().Trace(tgtURL), "Target ‘"+tgtURL+"’ is not a folder.")
		}
	}
}