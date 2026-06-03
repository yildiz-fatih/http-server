package main

import "html/template"

// parse templates at startup
var (
	errorTemplate      = template.Must(template.New("error").Parse(errorTemplateHTML))
	dirListingTemplate = template.Must(template.New("dirListing").Parse(dirListingTemplateHTML))
)

const errorTemplateHTML = `
<!DOCTYPE HTML>
<html lang="en">
   <head>
      <meta charset="utf-8">
      <style type="text/css">
         :root {
         color-scheme: light dark;
         }
      </style>
      <title>Error response</title>
   </head>
   <body>
      <h1>Error response</h1>
      <p>Error code: {{ .Code }}</p>
      <p>Message: {{ .Message }}</p>
   </body>
</html>
`

const dirListingTemplateHTML = `
<!DOCTYPE HTML>
<html lang="en">
   <head>
      <meta charset="utf-8">
      <style type="text/css">
         :root {
         color-scheme: light dark;
         }
      </style>
      <title>Directory listing for {{ .Path }}</title>
   </head>
   <body>
      <h1>Directory listing for {{ .Path }}</h1>
      <hr>
      <ul>
        {{ range .Files }}
         <li><a href="{{ . }}">{{ . }}</a></li>
        {{ end }}
      </ul>
      <hr>
   </body>
</html>
`
