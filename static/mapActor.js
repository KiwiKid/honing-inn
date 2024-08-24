


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
    }
  
    /**
     * Initializes the map using query string parameters for latitude, longitude, and zoom level.
     */
    initMap() {
      const queryParams = new URLSearchParams(window.location.search);
      const lat = parseFloat(queryParams.get('lat')) || -43.53937676715642;
      const lng = parseFloat(queryParams.get('lng')) || 172.55882263183597;
      const zoom = parseInt(queryParams.get('zoom'), 10) || 13;
  
      this.map = L.map(this.mapContainerId).setView([lat, lng], zoom);

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
      const marker = L.marker([lat, lng], options).addTo(this.map);
      this.markers.push(marker);
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

      const polygon = L.polygon(latlngs, areaOptions).bindPopup(`<div hx-get="/shapes/${areaOptions.shapeId}" hx-trigger="revealed">loading...</div>`, bindOptions).openPopup().addTo(this.map)      ;
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
        console.log(`updateUrlQuery ${JSON.stringify(details)}`)
        const queryString = new URLSearchParams(details).toString();
        console.log(`updateUrlQuery >>>>>  ${queryString}`)

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
    
      