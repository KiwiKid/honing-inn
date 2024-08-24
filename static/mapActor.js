


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
      
      this.createPopupOptions = {
        color: 'green',
        minWidth: '300'
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
    }
  
    /**
     * Initializes the map using query string parameters for latitude, longitude, and zoom level.
     */

    initMap() {
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
        
        var warning = L.layerGroup([], { collapsed: false});
        var noGo = L.layerGroup([]);
        var good = L.layerGroup([]);
        var homes = L.layerGroup([],  { collapsed: false});
        var redFlags = L.layerGroup([]);

        var overlayMaps = {
            "warning": warning,
            "noGo": noGo,
            "good": good,
            "homes": homes,
            "redFlags": redFlags
              
        };
      const queryParams = new URLSearchParams(window.location.search);
      const lat = parseFloat(queryParams.get('lat')) || -43.53937676715642;
      const lng = parseFloat(queryParams.get('lng')) || 172.55882263183597;
      const zoom = parseInt(queryParams.get('zoom'), 10) || 13;
  
      this.map = L.map(this.mapContainerId, {
        center: [lat, lng],
        zoom: zoom,
        layers: [osm]
      })


      var layerControl = L.control.layers(baseMaps, overlayMaps, { collapsed: false}).addTo(this.map);
      window.layerControl = layerControl
      window.overlayMaps = overlayMaps
      
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

      L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
      }).addTo(this.map);

      this.map.on('click', (e) => this.handleMapClick(e, this.map))
      this.map.on('moveend', (e) => this.handleMapMoveEnd(e))
      this.map.on('zoonend', (e) => this.handleMapMoveEnd(e))


      window.myMap = this.map;
      window.mapActor = this;

      this.addCustomControl();
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
         /*   const popup = `<div class="new-popup" hx-target="this">
                <form hx-post="/homes">
                    <input type="hidden" name="lat" value="${latlng.lat}"></input>
                    <input type="hidden" name="lng" value="${latlng.lng}"></input>
                    <button type="submit">Create Point</button>
                    <button onclick="location.reload();">cancel</button>
                </form>
             </div>`
*/

            const popup = `<div hx-get="/homes?lat=${latlng.lat}&lng=${latlng.lng}" hx-trigger="revealed">loading point..</div>`
           //map.addMarker(latlng.lat, latlng.lng, {title: 'New Point'}).bindPopup(popup).openPopup().addTo(map);

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
                
              /*  const popup = `<div class="new-popup" hx-target="this">
                <form hx-post="/shapes?updateMode=create-area">
                    <input type="hidden" name="shapeData" value="${JSON.stringify(window.existingNewAreaPoints)}"></input>
                    <input type="hidden" name=
                    <button type="submit">Save Area</button>
                </form>`*/

                   const popup = `<div hx-get="/shapes?mode=area" hx-trigger="revealed">loading..</div>`
                    window.leaflet.polygon(window.existingNewAreaPoints, {color: 'green'}).bindPopup(popup, this.createAreaPopupOptions).openPopup().addTo(map);
              //  window.leaflet.polygon(window.existingNewAreaPoints, {color: 'red'}).bindPopup(popup).openPopup().addTo(map);
            }
            break;
        }
        case '---': {
            highlightElement(document.getElementById('mode'))
        }
        default: {
            console.error(`mode  "${mode.value}" not found`)
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
        controlDiv.innerHTML = `
          <div class="tools" onclick="(e) => e.stopPropagation()">
            <select id="mode" class="modeset" name="mode" onclick="(e) => e.stopPropagation()">
              <option value="---">---</option>
              <option value="point">Create Points</option>
              <option value="area">Create Areas</option>
            </select>
          </div>
        `;

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
        throw new Error(`Invalid mode: ${newMode}`);
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
            html: `
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
            `,
            iconSize: [24, 24] // Adjust the size as needed
        });

          options.icon = flagIcon

          const flagMarker = L.marker([lat, lng], options).bindPopup(`<div hx-get="/homes${options.homeId ? `/${options.homeId}` :'' }" hx-trigger="revealed"></div>`, window.mapActor.editPointPopupOptions)

          overlayMaps.redFlags.addLayer(flagMarker)
          overlayMaps.redFlags.addTo(this.map)
          this.markers.push(flagMarker);

          break;
        case "Home":
          var houseIcon = L.divIcon({
            className: 'custom-icon',
            html: `
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
            `,
            iconSize: [24, 24] // Adjust the size as needed
        });
          options.icon = houseIcon
          const marker = L.marker([lat, lng], options).bindPopup(`<div hx-get="/homes${options.homeId ? `/${options.homeId}` :'' }" hx-trigger="revealed"></div>`, window.mapActor.editPointPopupOptions)

          overlayMaps.homes.addLayer(marker)
          overlayMaps.homes.addTo(this.map)
          this.markers.push(marker);

          break;
        default: 
        console.warn(`addMarker - pointKind "${options.pointKind}" not found`)
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
          console.error(`shapeKind ${areaOptions.shapeKind} not found`)
      } 

      const layerGroup = shapeLayerMapping[areaOptions.shapeKind] || L.layerGroup()

      const polygon = L.polygon(latlngs, areaOptions).bindPopup(`<div hx-get="/shapes/${areaOptions.shapeId}" hx-trigger="revealed">loading...</div>`, bindOptions).openPopup().addTo(this.map)          .addTo(layerGroup);

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


  
  }
















    function startArea(lat, lng){
        setMode('area')
        console.log(`startArea ${lat} ${lng}`)
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
        history.pushState(null, '', `?${queryString}`);
    }

    function addQueryParam(name, value) {

        const queryString = new URLSearchParams(window.location.search);
        queryString.set(name, value);

        const url = queryString.toString();  // Fixed the typo
        history.pushState(null, '', `?${url}`);
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
            controlDiv.innerHTML = `<div class="tools" click="(e) => e.stopPropagation()" >
            <select id="mode" class="modeset" name="mode" click="(e) => e.stopPropagation()" >
                        <option value="---">---</option>
                        <option value="point">Create Points</option>
                        <option value="area">Create Areas</option>
                    </select>
            </div>
            `;

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


    /** @@deprecated */
document.addEventListener("DOMContentLoaded", function() {
      // Create an instance of mapActor and initialize the map
      const mapController = new mapActor('map');
      mapController.initMap();
      
      mapController.setMode('point');
      

      // Clear markers and polygons (when needed)
      // mapController.clearMarkers();
      // mapController.clearPolygons();
    });


    // Add the custom control to the map


      const info = document.querySelector("#info");

      const setMode = function(mode){
            console.log(`setMode ${mode}`)
            mode.value = mode
      }
    
      