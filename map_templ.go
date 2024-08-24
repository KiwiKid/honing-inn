// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.747
package main

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"fmt"
)

func mapper(meta MapMeta, homes []Home, shapes []Shape) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<head><link rel=\"stylesheet\" href=\"https://unpkg.com/leaflet@1.9.3/dist/leaflet.css\" integrity=\"sha256-kLaT2GOSpHechhsozzB+flnD+zUyjE2LlfWPgU04xyI=\" crossorigin=\"\"><link rel=\"stylesheet\" href=\"reset.css\"><link rel=\"stylesheet\" href=\"app.css\"><script src=\"https://unpkg.com/leaflet@1.9.3/dist/leaflet.js\" integrity=\"sha256-WBkoXOwTeyKclOHuWtc+i2uENFpDZ9YPdf5Hf+D7ewM=\" crossorigin=\"\"></script><script src=\"https://unpkg.com/htmx.org@1.9.0\" integrity=\"sha384-aOxz9UdWG0yBiyrTwPeMibmaoq07/d3a96GCbb9x60f3mOt5zwkjdbcHFnKH8qls\" crossorigin=\"anonymous\"></script></head><body><!--<script src=\"./static/mapActor.js\"></script>--><style>\n    #map {\n        height: 100%;\n        width: 100%;\n    }\n\n\n.custom-control {\n  box-sizing: border-box;\n  background-color: #fff;\n  border: 1px solid #ccc;\n  line-height: 31px;\n  text-align: center;\n  text-decoration: none;\n  color: black;\n  border-radius: 2px;\n  width: 33px;\n  height: 33px;\n  border: 2px solid rgba(0, 0, 0, 0.2);\n}\n\n\n.modeset2 {\n  position: absolute;\n  top: 98px;\n  right: 10px;\n  z-index: 99999 !important;\n  font-size: 15px;\n  height: 30px;\n  width: 120px;\n}\n\n.infobox {\n    height: 200px;\n    width: 200px;\n}\n\n.tools { \n    position: absolute;\n  top: 98px;\n  right: 10px;\n    z-index: 99999 !important;\npadding: 10px;\n  height: 30px;\n  width: 120px;\n    \n}\n\n\n/* Basic form input styles */\ninput, textarea, select, button {\n  padding: 0.75rem 1rem; /* Increase padding */\n  font-size: 1rem; /* Slightly larger font */\n  border-radius: 0.375rem; /* Tailwind's rounded-md equivalent */\n  border: 1px solid #d1d5db; /* Light border */\n  background-color: #f9fafb; /* Light background */\n  transition: border-color 0.2s ease, box-shadow 0.2s ease; /* Smooth transitions */\n}\n\n/* Focus state for better accessibility */\ninput:focus, textarea:focus, select:focus, button:focus {\n  outline: none;\n  border-color: #3b82f6; /* Tailwind blue-500 */\n  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.5); /* Focus shadow */\n}\n\n/* Basic button styles */\nbutton {\n  background-color: #3b82f6; /* Tailwind blue-500 */\n  color: white;\n  transition: background-color 0.2s ease;\n}\n\n/* Button hover state */\nbutton:hover {\n  background-color: #2563eb; /* Darker blue */\n}\n\n/* Responsive adjustments for smaller screens */\n@media (max-width: 640px) {\n  input, textarea, select, button {\n    width: 100%; /* Full width on small screens */\n  }\n}\n\n\n    </style><div id=\"map\" class=\"map\" data-center=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("[%f, %f]", meta.Lat, meta.Lng))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `map.templ`, Line: 119, Col: 88}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" data-zoom=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("%d", meta.Zoom))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `map.templ`, Line: 119, Col: 131}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><div data-tile=\"EsriWorldImagery\" data-max-zoom=\"19\" data-min-zoom=\"5\" data-default></div><div data-tile=\"OpenStreetMap\"></div><table id=\"map-container\"><div id=\"info\" class=\"leaflet-bar leaflet-control infobox\"><span class=\" zoom-level\" id=\"zoom-level\"></span></div><div hx-get=\"/shapes?mode=all\" hx-trigger=\"revealed\">loading shapes..</div></table></div><script src=\"/static/mapActor.js\" type=\"text/javascript\"></script></body>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}
