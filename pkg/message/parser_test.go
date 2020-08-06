// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package message

import (
	"bytes"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
)

func TestParseTextPlain(t *testing.T) {
	f := f("text_plain.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextPlainUTF8(t *testing.T) {
	f := f("text_plain_utf8.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextPlainLatin1(t *testing.T) {
	f := f("text_plain_latin1.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "ééééééé", m.Body)
	assert.Equal(t, "ééééééé", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextPlainUnknownCharsetIsActuallyLatin1(t *testing.T) {
	f := f("text_plain_unknown_latin1.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "ééééééé", m.Body)
	assert.Equal(t, "ééééééé", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextPlainUnknownCharsetIsActuallyLatin2(t *testing.T) {
	f := f("text_plain_unknown_latin2.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// The file contains latin2-encoded text, but we will assume it is latin1
	// and decode it as such. This will lead to corruption.
	latin2, _ := charmap.ISO8859_2.NewEncoder().Bytes([]byte("řšřšřš"))
	expect, _ := charmap.ISO8859_1.NewDecoder().Bytes(latin2)
	assert.NotEqual(t, []byte("řšřšřš"), expect)

	assert.Equal(t, string(expect), m.Body)
	assert.Equal(t, string(expect), plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextPlainAlready7Bit(t *testing.T) {
	f := f("text_plain_7bit.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextPlainWithOctetAttachment(t *testing.T) {
	f := f("text_plain_octet_attachment.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	require.Len(t, attReaders, 1)
	assert.Equal(t, readerToString(attReaders[0]), "if you are reading this, hi!")
}

func TestParseTextPlainWithOctetAttachmentGoodFilename(t *testing.T) {
	f := f("text_plain_octet_attachment_good_2231_filename.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	assert.Len(t, attReaders, 1)
	assert.Equal(t, readerToString(attReaders[0]), "if you are reading this, hi!")
	assert.Equal(t, "😁😂.txt", m.Attachments[0].Name)
}

func TestParseTextPlainWithOctetAttachmentBadFilename(t *testing.T) {
	f := f("text_plain_octet_attachment_bad_2231_filename.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	assert.Len(t, attReaders, 1)
	assert.Equal(t, readerToString(attReaders[0]), "if you are reading this, hi!")
	assert.Equal(t, "attachment.bin", m.Attachments[0].Name)
}

func TestParseTextPlainWithPlainAttachment(t *testing.T) {
	f := f("text_plain_plain_attachment.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	require.Len(t, attReaders, 1)
	assert.Equal(t, readerToString(attReaders[0]), "attachment")
}

func TestParseTextPlainWithImageInline(t *testing.T) {
	f := f("text_plain_image_inline.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	// The inline image is an 8x8 mic-dropping gopher.
	require.Len(t, attReaders, 1)
	img, err := png.DecodeConfig(attReaders[0])
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseWithMultipleTextParts(t *testing.T) {
	f := f("multiple_text_parts.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body\nsome other part of the message", m.Body)
	assert.Equal(t, "body\nsome other part of the message", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextHTML(t *testing.T) {
	f := f("text_html.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> without attachment</body></html>", m.Body)
	assert.Equal(t, "This is body of *HTML mail* without attachment", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextHTMLAlready7Bit(t *testing.T) {
	f := f("text_html_7bit.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> without attachment</body></html>", m.Body)
	assert.Equal(t, "This is body of *HTML mail* without attachment", plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseTextHTMLWithOctetAttachment(t *testing.T) {
	f := f("text_html_octet_attachment.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> with attachment</body></html>", m.Body)
	assert.Equal(t, "This is body of *HTML mail* with attachment", plainBody)

	require.Len(t, attReaders, 1)
	assert.Equal(t, readerToString(attReaders[0]), "if you are reading this, hi!")
}

func TestParseTextHTMLWithPlainAttachment(t *testing.T) {
	f := f("text_html_plain_attachment.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// BAD: plainBody should not be empty!
	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> with attachment</body></html>", m.Body)
	assert.Equal(t, "This is body of *HTML mail* with attachment", plainBody)

	require.Len(t, attReaders, 1)
	assert.Equal(t, readerToString(attReaders[0]), "attachment")
}

func TestParseTextHTMLWithImageInline(t *testing.T) {
	f := f("text_html_image_inline.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	assert.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "<html><body>This is body of <b>HTML mail</b> with attachment</body></html>", m.Body)
	assert.Equal(t, "This is body of *HTML mail* with attachment", plainBody)

	// The inline image is an 8x8 mic-dropping gopher.
	require.Len(t, attReaders, 1)
	img, err := png.DecodeConfig(attReaders[0])
	require.NoError(t, err)
	assert.Equal(t, 8, img.Width)
	assert.Equal(t, 8, img.Height)
}

func TestParseWithAttachedPublicKey(t *testing.T) {
	f := f("text_plain.eml")

	// BAD: Public Key is not attached unless Content-Type is specified (not required)!
	m, plainBody, attReaders, err := Parse(f, "publickey", "publickeyname")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	assert.Equal(t, "body", m.Body)
	assert.Equal(t, "body", plainBody)

	// HELP: Should public key be available as an attachment? In previous parser it wasn't...
	require.Len(t, attReaders, 1)
}

func TestParseTextHTMLWithEmbeddedForeignEncoding(t *testing.T) {
	f := f("text_html_embedded_foreign_encoding.eml")

	m, plainBody, attReaders, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"Sender" <sender@pm.me>`, m.Sender.String())
	assert.Equal(t, `"Receiver" <receiver@pm.me>`, m.ToList[0].String())

	// BAD: Bridge does not detect the charset specified in the <meta> tag of the html.
	assert.Equal(t, `<html><head><meta charset="ISO-8859-2"></head><body>latin2 řšřš</body></html>`, m.Body)
	assert.Equal(t, `latin2 řšřš`, plainBody)

	assert.Len(t, attReaders, 0)
}

func TestParseMultipartAlternative(t *testing.T) {
	f := f("multipart_alternative.eml")

	m, plainBody, _, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"schizofrenic" <schizofrenic@pm.me>`, m.Sender.String())
	assert.Equal(t, `<pmbridgeietest@outlook.com>`, m.ToList[0].String())

	assert.Equal(t, `<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    <b>aoeuaoeu</b>
  </body>
</html>
`, m.Body)

	assert.Equal(t, "*aoeuaoeu*\n\n", plainBody)
}

func TestParseMultipartAlternativeNested(t *testing.T) {
	f := f("multipart_alternative_nested.eml")

	m, plainBody, _, err := Parse(f, "", "")
	require.NoError(t, err)

	assert.Equal(t, `"schizofrenic" <schizofrenic@pm.me>`, m.Sender.String())
	assert.Equal(t, `<pmbridgeietest@outlook.com>`, m.ToList[0].String())

	assert.Equal(t, `<html>
  <head>
    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    <b>multipart 2.2</b>
  </body>
</html>
`, m.Body)

	assert.Equal(t, "*multipart 2.1*\n\n", plainBody)
}

func f(filename string) []byte {
	f, err := os.Open(filepath.Join("testdata", filename))
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)

	_, _ = buf.ReadFrom(f)

	return buf.Bytes()
}

func readerToString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)

	if err != nil {
		panic(err)
	}

	return string(b)
}
