/*
 *
 *  Copyright 2019 Tero Vierimaa
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *
 */

package models

import (
	"net/url"
	"strings"
	"time"
)

type Bookmark struct {
	Id          int
	Name        string
	LowerName   string
	Description string
	Content     string
	Project     string
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	Tags []string
}

//Return domain of the content if it is a link
func (b *Bookmark) ContentDomain() string {

	// url.prase rarely gives any error, however invalid domain isn't parsed and returns ""
	Url, err := url.Parse(b.Content)
	if err != nil {
		return "not url"
	}

	return Url.Host
}

//TagsString retuns string representation of tags.
//If spaces flag is set, put comma and space between tags
// No tags -> "", tags -> "a, b"
func (b *Bookmark) TagsString(spaces bool) string {
	if len(b.Tags) == 0 {
		return ""
	}

	separator := ","
	if spaces {
		separator += " "
	}
	return strings.Join(b.Tags, separator)
}
