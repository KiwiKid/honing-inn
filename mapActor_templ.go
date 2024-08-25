// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.747
package main

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func span() templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<span></span>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

func mapActor() templ.ComponentScript {
	return templ.ComponentScript{
		Name: `__templ_mapActor_40aa`,
		Function: `function __templ_mapActor_40aa(){/**
 * @typedef {import('https://cdn.jsdelivr.net/npm/@types/leaflet/index.d.ts').Map} L 
 * @typedef {import('https://cdn.jsdelivr.net/npm/@types/leaflet/index.d.ts').Marker} L.Marker
 * @typedef {import('https://cdn.jsdelivr.net/npm/@types/leaflet/index.d.ts').LatLng} L.LatLng
 */


console.log('mapActor')
if(window.mapActor){
  console.log('map actor ran - already init');
   return;
}


class mapActor {

    /**
     * @param {string} mapContainerId - The ID of the HTML element where the map will be rendered.
     * @param {'none'|'point'|'area'} [initialMode='none'] - The initial mode of interaction with the map.
     */
    constructor(mapContainerId, initialMode = 'none') {
      this.map = null;
      this.markers = [];
      this.polygons = [];
      this.mode = initialMode;
      this.mapContainerId = mapContainerId;
      this.mapMeta = null

      
      this.createPopupOptions = {
        color: 'green',
        minWidth: '500'
      }

      this.createAreaPopupOptions = {
        color: 'red',
        minWidth: '300'
      }

      this.editAreaPopupOptions = {
        color: 'blue',
        minWidth: '300'
      }

      this.editPointPopupOptions = {
        color: 'blue',
        minWidth: '300'
      }


      window.shapeLayers = {};

      window.mapActor = this;
      window.myMap = this.map;

      var warning = L.layerGroup([], { collapsed: false});
      var noGo = L.layerGroup([]);
      var good = L.layerGroup([]);
      var homes = L.layerGroup([],  { collapsed: false});
      var redFlags = L.layerGroup([]);

      window.overlayMaps = {
          "warning": warning,
          "noGo": noGo,
          "good": good,
          "homes": homes,
          "redFlags": redFlags
      };

    }
  
    /**
     * Initializes the map using query string parameters for latitude, longitude, and zoom level.
     */ 

    initMap() {
      this.mapMeta = JSON.parse(document.getElementById(this.mapContainerId).getAttribute('data-meta')) 

      var osm = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '© OpenStreetMap'
      });

      var osmHOT = L.tileLayer('https://{s}.tile.openstreetmap.fr/hot/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '© OpenStreetMap contributors, Tiles style by Humanitarian OpenStreetMap Team hosted by OpenStreetMap France'});
    

        var baseMaps = {
          "OpenStreetMap": osm,
          "OpenStreetMap.HOT": osmHOT
      };

      if(this.mapMeta.ProcessMode){
        this.map = L.map(this.mapContainerId, {
          center: [this.mapMeta.Lat, this.mapMeta.Lng],
          zoom: this.mapMeta.Zoom,
          layers: [osm, osmHOT]
        })

        window.myMap = this.map;
        window.mapActor = this;
        const mapProcessingMeta = JSON.parse(document.getElementById(this.mapContainerId).getAttribute('data-processing-meta'))

        const ProcessResults = L.Control.extend({
            options: {
                position: 'topright'
            },

            onAdd: function(map) {
                // Create a div element with the 'custom-control' class
                const controlDiv = L.DomUtil.create('div', 'custom-control');

                // Add content to the control
                controlDiv.innerHTML = ` + "`" + `<div id="images">
                </div>
                ` + "`" + `;


                

                return controlDiv;
            },

            onRemove: function(map) {
                // Nothing to clean up when the control is removed
            }
        });


        this.map.addControl(new ProcessResults());

        this.handleMapProcessing(mapProcessingMeta)

        return
      }
      
      this.map = L.map(this.mapContainerId, {
        center: [this.mapMeta.Lat, this.mapMeta.Lng],
        zoom: this.mapMeta.Zoom,
        layers: [osm, osmHOT],
        preferCanvas: true
      })


      var layerControl = L.control.layers(baseMaps, overlayMaps, { collapsed: false}).addTo(this.map);
      window.layerControl = layerControl
=      
        this.map.on('popupopen', function(e) {
              const popups = document.querySelectorAll(".leaflet-popup-content-wrapper")
              popups.forEach((p) => {
                  htmx.process(p)
              })
              const popup = e.popup; // e.popup refers to the currently opened popup

        // Update the popup to refresh its size
        if (popup) {
   //         popup.update();
        }
          });

   /*   L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
      }).addTo(this.map);*/

      this.map.on('click', (e) => this.handleMapClick(e, this.map))
      this.map.on('moveend', (e) => this.handleMapMoveEnd(e))
      this.map.on('zoonend', (e) => this.handleMapMoveEnd(e))



      this.addCustomControl();
      this.processShapesAndHomes();
  }

handleMapClick(event, map){
    const { latlng } = event;
    
    if(event.target.name === "mode"){
        console.log('non-map click')
        console.log(event)
        return    
    }else {
        console.log('map click')
        console.log(event)
    }

    
    let mode = document.getElementById('mode')

    switch(mode.value){
        case 'point':{

            const popup = ` + "`" + `<div hx-get="/homes?lat=${latlng.lat}&lng=${latlng.lng}" hx-trigger="revealed">loading point..</div>` + "`" + `

            const marker = window.leaflet.circleMarker([latlng.lat, latlng.lng], { color: 'green', radius: 10})
 
             marker.addTo(map).bindPopup(popup, this.createPopupOptions)
             .openPopup();;


            const popups = document.querySelectorAll(".new-popup")
            popups.forEach((p) => {
                htmx.process(p)
            })
            break;
        }
        case 'area':{

            if(!window.existingNewAreaPoints){
                window.existingNewAreaPoints = [[latlng.lat, latlng.lng]]
                const polyGon = window.leaflet.polygon(window.existingNewAreaPoints, {color: 'green'}).addTo(map);        
                window.existingNewArea = polyGon  
            }else {
                window.existingNewAreaPoints = [...window.existingNewAreaPoints, [latlng.lat, latlng.lng]]
                if(window.existingNewArea){
                 //   window.existingNewArea.remove()           
                }

                   const popup = ` + "`" + `<div hx-get="/shapes?mode=area" hx-trigger="revealed">loading..</div>` + "`" + `
                    window.leaflet.polygon(window.existingNewAreaPoints, {color: 'green'}).bindPopup(popup, this.createAreaPopupOptions).openPopup().addTo(map);
              //  window.leaflet.polygon(window.existingNewAreaPoints, {color: 'red'}).bindPopup(popup).openPopup().addTo(map);
            }
            break;
        }
        case '---': {
            highlightElement(document.getElementById('mode'))
        }
        default: {
            console.error(` + "`" + `mode  "${mode.value}" not found` + "`" + `)
        }
    }
}

handleMapMoveEnd(e){
  const mode = document.getElementById('mode')
  const center = e.target.getCenter()
  debouncedUpdateUrlQuery({
    zoom: e.target._zoom,
    lat: center.lat,
    lng: center.lng,
    mode: mode.value
});
}

 handleMapZoomEnd(e){
  const mode = document.getElementById('mode')
  const center = e.target.getCenter()

  debouncedUpdateUrlQuery({
      zoom: e.target._zoom,
      lat: center.lat,
      lng: center.lng,
      mode: mode.value
  });
}




   
        


  addCustomControl() {
    const CustomControl = L.Control.extend({
      options: {
        position: 'topright'
      },

      onAdd: (map) => {
        // Create a div element with the 'custom-control' class
        const controlDiv = L.DomUtil.create('div', 'custom-control');

        // Add content to the control
        controlDiv.innerHTML = ` + "`" + `
          <div class="tools" onclick="(e) => e.stopPropagation()">
            <select id="mode" class="modeset" name="mode" onclick="(e) => e.stopPropagation()">
              <option value="---">---</option>
              <option value="point">Create Points</option>
              <option value="area">Create Areas</option>
            </select>
          </div>
        ` + "`" + `;

        // Set the select value based on the current query parameter
        const modeSetting = new URLSearchParams(window.location.search).get('mode') || '---';
        const select = controlDiv.querySelector('#mode');
        select.value = modeSetting;

        // Add event listener for changes in mode
        select.addEventListener('change', (event) => {
          const selectedMode = event.target.value;
          this.setMode(selectedMode === '---' ? 'none' : selectedMode);
          this.addQueryParam('mode', selectedMode);
        });

        return controlDiv;
      },

      onRemove: function(map) {
        // Nothing to clean up when the control is removed
      }
    });

    // Add the custom control to the map
    this.map.addControl(new CustomControl());
  }
  
    /**
     * Sets the mode of interaction with the map.
     * @param {'none'|'point'|'area'} newMode - The new mode of interaction with the map.
     * @throws {Error} If the mode is invalid.
     */
    setMode(newMode) {
      if (['none', 'point', 'area'].includes(newMode)) {
        this.mode = newMode;
      } else {
        throw new Error(` + "`" + `Invalid mode: ${newMode}` + "`" + `);
      }
    }
  
    /**
     * Adds a marker to the map.
     * @param {number} lat - Latitude of the marker.
     * @param {number} lng - Longitude of the marker.
     * @param {L.MarkerOptions} [options={}] - Optional settings for the marker.
     */
    addMarker(lat, lng, options = {}) {

     
      
      switch(options.pointKind){
        
        case "RedFlag":

          var flagIcon = L.divIcon({
            className: 'custom-icon',
            html: ` + "`" + `
            <svg height="2rem" width="2rem" version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" 
            viewBox="0 0 506.4 506.4" xml:space="preserve">
         <circle style="fill:#DF5C4E;" cx="253.2" cy="253.2" r="249.2"/>
         <g>
           <path style="fill:#F4EFEF;" d="M253.2,332.4c-10.8,0-20-8.8-20-19.6v-174c0-10.8,9.2-19.6,20-19.6s20,8.8,20,19.6v174
             C273.2,323.6,264,332.4,253.2,332.4z"/>
           <path style="fill:#F4EFEF;" d="M253.2,395.6c-5.2,0-10.4-2-14-5.6s-5.6-8.8-5.6-14s2-10.4,5.6-14s8.8-6,14-6s10.4,2,14,6
             c3.6,3.6,6,8.8,6,14s-2,10.4-6,14C263.6,393.6,258.4,395.6,253.2,395.6z"/>
         </g>
         <path d="M253.2,506.4C113.6,506.4,0,392.8,0,253.2S113.6,0,253.2,0s253.2,113.6,253.2,253.2S392.8,506.4,253.2,506.4z M253.2,8
           C118,8,8,118,8,253.2s110,245.2,245.2,245.2s245.2-110,245.2-245.2S388.4,8,253.2,8z"/>
         <path d="M249.2,336.4c-13.2,0-24-10.8-24-23.6v-174c0-13.2,10.8-23.6,24-23.6s24,10.8,24,23.6v174
           C273.2,325.6,262.4,336.4,249.2,336.4z M249.2,122.8c-8.8,0-16,7.2-16,15.6v174c0,8.8,7.2,15.6,16,15.6s16-7.2,16-15.6v-174
           C265.2,130,258,122.8,249.2,122.8z"/>
         <path d="M249.2,399.6c-6.4,0-12.4-2.4-16.8-6.8c-4.4-4.4-6.8-10.4-6.8-16.8s2.4-12.4,6.8-16.8c4.4-4.4,10.8-6.8,16.8-6.8
           c6.4,0,12.4,2.4,16.8,6.8c4.4,4.4,6.8,10.4,6.8,16.8s-2.4,12.4-7.2,16.8C261.6,397.2,255.6,399.6,249.2,399.6z M249.2,360
           c-4,0-8.4,1.6-11.2,4.8c-2.8,2.8-4.4,6.8-4.4,11.2c0,4,1.6,8.4,4.8,11.2c2.8,2.8,7.2,4.8,11.2,4.8s8.4-1.6,11.2-4.8
           c2.8-2.8,4.8-7.2,4.8-11.2s-1.6-8.4-4.8-11.2C257.2,361.6,253.2,360,249.2,360z"/>
         </svg>
            ` + "`" + `,
            iconSize: [24, 24] // Adjust the size as needed
        });

          options.icon = flagIcon

          const flagMarker = L.marker([lat, lng], options).bindPopup(` + "`" + `<div hx-get="/homes${options.homeId ? ` + "`" + `/${options.homeId}` + "`" + ` :'' }" hx-trigger="revealed"></div>` + "`" + `, window.mapActor.editPointPopupOptions)

          overlayMaps.redFlags.addLayer(flagMarker)
          overlayMaps.redFlags.addTo(this.map)
          this.markers.push(flagMarker);

          break;
        case "Home":
          var houseIcon = L.divIcon({
            className: 'custom-icon',
            html: ` + "`" + `
            <svg fill="#000000" height="2rem" width="2rem" version="1.1" id="Capa_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" 
            viewBox="0 0 486.196 486.196" xml:space="preserve">
         <g>
           <path d="M481.708,220.456l-228.8-204.6c-0.4-0.4-0.8-0.7-1.3-1c-5-4.8-13-5-18.3-0.3l-228.8,204.6c-5.6,5-6,13.5-1.1,19.1
             c2.7,3,6.4,4.5,10.1,4.5c3.2,0,6.4-1.1,9-3.4l41.2-36.9v7.2v106.8v124.6c0,18.7,15.2,34,34,34c0.3,0,0.5,0,0.8,0s0.5,0,0.8,0h70.6
             c17.6,0,31.9-14.3,31.9-31.9v-121.3c0-2.7,2.2-4.9,4.9-4.9h72.9c2.7,0,4.9,2.2,4.9,4.9v121.3c0,17.6,14.3,31.9,31.9,31.9h72.2
             c19,0,34-18.7,34-42.6v-111.2v-34v-83.5l41.2,36.9c2.6,2.3,5.8,3.4,9,3.4c3.7,0,7.4-1.5,10.1-4.5
             C487.708,233.956,487.208,225.456,481.708,220.456z M395.508,287.156v34v111.1c0,9.7-4.8,15.6-7,15.6h-72.2c-2.7,0-4.9-2.2-4.9-4.9
             v-121.1c0-17.6-14.3-31.9-31.9-31.9h-72.9c-17.6,0-31.9,14.3-31.9,31.9v121.3c0,2.7-2.2,4.9-4.9,4.9h-70.6c-0.3,0-0.5,0-0.8,0
             s-0.5,0-0.8,0c-3.8,0-7-3.1-7-7v-124.7v-106.8v-31.3l151.8-135.6l153.1,136.9L395.508,287.156L395.508,287.156z"/>
         </g>
         </svg>
            ` + "`" + `,
            iconSize: [24, 24] // Adjust the size as needed
        });
          options.icon = houseIcon
          const marker = L.marker([lat, lng], options).bindPopup(` + "`" + `<div hx-get="/homes${options.homeId ? ` + "`" + `/${options.homeId}` + "`" + ` :'' }?viewMode=view" hx-trigger="revealed"></div>` + "`" + `, window.mapActor.editPointPopupOptions)

          overlayMaps.homes.addLayer(marker)
          overlayMaps.homes.addTo(this.map)
          this.markers.push(marker);

          break;
        default: 
        console.warn(` + "`" + `addMarker - pointKind "${options.pointKind} not found` + "`" + `)
      }

    }
  
    /**
     * Adds a polygon to the map.
     * @param {Array<Array<number>>} latlngs - An array of latitude and longitude pairs defining the polygon's vertices.
     * @param {L.PolylineOptions} [options={}] - Optional settings for the polygon.
     */
    addPolygon(latlngs, areaOptions = {}, bindOptions = { minWidth: '200px'}) {
      if(!latlngs){
        console.error('latlngs is required for addPolygon')
        return
      }else{
          if(typeof latlngs == "string"){
              latlngs = JSON.parse(latlngs)
          }
      }

      const shapeLayerMapping = {
        "warning": overlayMaps.warning,
        "noGo": overlayMaps.noGo,
        "good": overlayMaps.good,
        "homes": overlayMaps.homes,
        "redFlags": overlayMaps.redFlags
    };
      
      switch(areaOptions.shapeKind){
        case "good":
          areaOptions.color = '#169016'
          break;
        case "no-go":
          areaOptions.color = 'black'
          break
        case "warning":
          areaOptions.color = 'red'
          break;
        
        default:
          console.error(` + "`" + `shapeKind ${areaOptions.shapeKind} not found` + "`" + `)
      } 

      const layerGroup = shapeLayerMapping[areaOptions.shapeKind] || L.layerGroup()

      const polygon = L.polygon(latlngs, areaOptions).bindPopup(` + "`" + `<div hx-get="/shapes/${areaOptions.shapeId}" hx-trigger="revealed">loading...</div>` + "`" + `, bindOptions).openPopup().addTo(this.map)          .addTo(layerGroup);

      layerGroup.addTo(this.map)
      
      
      this.polygons.push(polygon);
    }
  
    /**
     * Removes all markers from the map.
     */
    clearMarkers() {
      this.markers.forEach(marker => this.map.removeLayer(marker));
      this.markers = [];
    }
  
    /**
     * Removes all polygons from the map.
     */
    clearPolygons() {
      this.polygons.forEach(polygon => this.map.removeLayer(polygon));
      this.polygons = [];
    }
  
    /**
     * Adds markers to the map based on backend data.
     * @param {Array<{lat: number, lng: number, options?: L.MarkerOptions}>} data - An array of marker data from the backend.
     */
    addMarkersFromData(data) {
      data.forEach(item => this.addMarker(item.lat, item.lng, item.options));
    }

    addQueryParam(key, value) {
        const url = new URL(window.location);
        url.searchParams.set(key, value);
        window.history.replaceState({}, '', url);
    }



        // Function to collect and process shapes and homes
    processShapesAndHomes() {

          console.log('shape list js - processShapesAndHomes')
          // Group shapes into layers
          window.shapeLayers = {};

              console.log(' Add home markers')

         document.querySelectorAll('span[data-shape-id]').forEach(function(element) {
              console.log('Processing shapes');
              if (element.getAttribute('rendered') !== 'true') {
                  const shapeData = JSON.parse(element.getAttribute('data-shape-data'));
                  const shapeId = element.getAttribute('data-shape-id');
                  const shapeKind = element.getAttribute('data-shape-kind');

                  // Get the corresponding layer group for the shape kind
            
                  // Add the polygon to the layer group
                  window.mapActor.addPolygon(shapeData, { shapeKind: shapeKind, shapeId: shapeId }, window.mapActor.editAreaPopupOptions)
                     

                  element.setAttribute('rendered', 'true');
              }
          }.bind(this)); // Bind 'this' to ensure 'this.map' is accessible

          console.log('Adding home markers');

          // Add home markers
          document.querySelectorAll('span[data-home-id]').forEach(function(element) {
              console.log('Processing homes');
              if (element.getAttribute('rendered') !== 'true') {
                  const home = JSON.parse(element.getAttribute('data-home'))
                  console.log(home)
                  const lat = home.Lat;
                  const lng = home.Lng;
                  const homeId = home.ID;
                  const pointKind = home.PointType || "Home"

                  window.mapActor.addMarker(lat, lng, { homeId, pointKind })

                  element.setAttribute('rendered', 'true');
              }
          });
      }

      // Initialize map and process shapes and homes
   
      handleMapProcessing(mapProcessingMeta) {
        console.log('handleMapProcessing', mapProcessingMeta);
      
        const map = this.map;  // Reference to your Leaflet map instance
        const gridHeight = mapProcessingMeta.GridHeight;
        const gridWidth = mapProcessingMeta.GridWidth;
        let currentRow = 0;
        let currentCol = 0;
      
        // Move the map to the initial position
      //  map.setView(mapProcessingMeta.start_point, mapProcessingMeta.zoom);
      
        // Function to capture a screenshot and send it to the backend
        const processCurrentView = () => {
          this.captureScreenshot().then((imageData) => {

            const payload = {
              processingMeta: mapProcessingMeta,
              current_row: currentRow,
              current_col: currentCol,
              image_data: imageData,  // Base64 encoded image data
            };

            console.log(` + "`" + `FETCH /process ${JSON.stringify(payload)}` + "`" + `)
            fetch('/process', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json', // Specify the content type
              },
              body: JSON.stringify(payload), // Convert the payload object to a JSON string
            })
              .then(response => {
                if (!response.ok) {
                  throw new Error('Network response was not ok');
                }
                console.log(` + "`" + `processed OK ${JSON.stringify(response.body, null, 4)}` + "`" + `)
              })
              .then(data => {
                console.log(` + "`" + `Processed row ${currentRow}, col ${currentCol}: ` + "`" + `, data);
                moveToNextCell();
              })
              .catch(error => {

                console.error(` + "`" + `Error processing row ${currentRow}, col ${currentCol}: ` + "`" + `, error);
              });
          });
        };
      
        // Function to move to the next grid cell
        const moveToNextCell = () => {
          if (currentCol < gridWidth - 1) {
            currentCol++;
          } else {
            currentCol = 0;
            if (currentRow < gridHeight - 1) {
              currentRow++;
            } else {
              console.log("Grid processing complete");
              return;  // Processing is complete
            }
          }
      
          // Calculate the new center for the next grid cell based on map bounds
          const bounds = map.getBounds();
          const latDelta = bounds.getNorth() - bounds.getSouth();
          const lngDelta = bounds.getEast() - bounds.getWest();
      
          const newLat = bounds.getSouth() + currentRow * latDelta;
          const newLng = bounds.getWest() + currentCol * lngDelta;
      
          map.setView([newLat, newLng], mapProcessingMeta.zoom);  // Trigger the 'moveend' event
        };
      
        // Event listener for when the map finishes moving
        map.on('moveend', function (e) {
          // Use the event data to update the URL query parameter
      
          processCurrentView();  // Capture screenshot and process current view
        });
      
        // Start processing by moving to the first grid cell
        moveToNextCell();
      }
      
      // Example method to capture a screenshot of the map using leaflet-image
      captureScreenshot() {
        return new Promise((resolve, reject) => {
          leafletImage(this.map, function (err, canvas) {
            if (err) {
              reject(err);
              return;
            }
            
            
            // Convert the canvas to base64 image data
            const base64data = canvas.toDataURL().split(',')[1];  // Strip the data URL prefix
            resolve(base64data);  // Resolve with base64 encoded image

            var img = document.createElement('img');
            var dimensions = window.myMap.getSize();
            img.width = dimensions.x;
            img.height = dimensions.y;
            img.src = canvas.toDataURL();
            document.getElementById('images').innerHTML = '';
            document.getElementById('images').appendChild(img);
          });
        });
      }
      
  }

  console.log('INIT MAP')
  const mapController = new mapActor('map');
  mapController.initMap();

    function startArea(lat, lng){
        setMode('area')
        console.log(` + "`" + `startArea ${lat} ${lng}` + "`" + `)
    }

    function debounce(func, wait) {
        let timeout;
        return function (...args) {
            clearTimeout(timeout);
            timeout = setTimeout(() => func.apply(this, args), wait);
        };
    }

    // Function to update the URL query
    function updateUrlQuery(details) {
        const queryString = new URLSearchParams(details).toString();
        history.pushState(null, '', ` + "`" + `?${queryString}` + "`" + `);
    }

    function addQueryParam(name, value) {

        const queryString = new URLSearchParams(window.location.search);
        queryString.set(name, value);

        const url = queryString.toString();  // Fixed the typo
        history.pushState(null, '', ` + "`" + `?${url}` + "`" + `);
    }

    const debouncedUpdateUrlQuery = debounce(updateUrlQuery, 300);
    
    function highlightElement(element) {
        // Apply the highlight styles
        element.style.border = '5px solid yellow';
        
        // Start fading out the border after a delay (e.g., 2 seconds)
        setTimeout(() => {
            element.style.transition = 'border 2s ease-out';
            element.style.border = '0 solid transparent';
            
            // Optionally, reset the transition after it's done for repeatability
            setTimeout(() => {
            element.style.transition = '';
            }, 2000);  // Match the fade-out duration
        }, 2000);  // Duration of the highlight before fading out
    }

     // Define the custom control
    const CustomControl = L.Control.extend({
        options: {
            position: 'topright'
        },

        onAdd: function(map) {
            // Create a div element with the 'custom-control' class
            const controlDiv = L.DomUtil.create('div', 'custom-control');

            // Add content to the control
            controlDiv.innerHTML = ` + "`" + `<div class="tools" click="(e) => e.stopPropagation()" >
            <select id="mode" class="modeset" name="mode" click="(e) => e.stopPropagation()" >
                        <option value="---">---</option>
                        <option value="point">Create Points</option>
                        <option value="area">Create Areas</option>
                    </select>
            </div>
            ` + "`" + `;

            const modeSetting = new URLSearchParams(window.location.search)
            const select = controlDiv.firstChild
            select.value = modeSetting.get('mode')

            select.addEventListener('change', (elm) => {
                addQueryParam('mode', elm.target.value)
            })

            

            return controlDiv;
        },

        onRemove: function(map) {
            // Nothing to clean up when the control is removed
        }
    });
/*
    var maps = [];

    L.Map.addInitHook(function () {
        if(maps.length == 0){
            maps.push(this);
        }
        

        

       this.on('moveend', function(e){
                console.log('moveend')
        })
        this.on('move', function(e){
                console.log('move')
        })

      

        this.addControl(new CustomControl());

    });*/



      const info = document.querySelector("#info");

      const setMode = function(mode){
            console.log(` + "`" + `setMode ${mode}` + "`" + `)
            mode.value = mode
      }
    

      
}`,
		Call:       templ.SafeScript(`__templ_mapActor_40aa`),
		CallInline: templ.SafeScriptInline(`__templ_mapActor_40aa`),
	}
}
