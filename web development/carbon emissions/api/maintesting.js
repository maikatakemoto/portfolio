let map, autocomplete1, autocomplete2, directionsService, directionsDisplay;

function initMap() {
    var input1 = document.getElementById("from");
    var input2 = document.getElementById("to");

    autocomplete1 = new google.maps.places.Autocomplete(input1);
    autocomplete2 = new google.maps.places.Autocomplete(input2);

    var myLatLng = {
        lat: 38.3460,
        lng: -0.4907
    };

    var mapOptions = {
        center: myLatLng,
        zoom: 7,
        mapTypeId: google.maps.MapTypeId.ROADMAP
    };

    map = new google.maps.Map(document.getElementById("map"), mapOptions);

    directionsService = new google.maps.DirectionsService();
    directionsDisplay = new google.maps.DirectionsRenderer();
    directionsDisplay.setMap(map);
}

function calcRoute() {
    var request = {
        origin: document.getElementById("from").value,
        destination: document.getElementById("to").value,
        travelMode: google.maps.TravelMode.DRIVING,
        unitSystem: google.maps.UnitSystem.IMPERIAL
    }

    directionsService.route(request, function(result, status) {
        if (status == google.maps.DirectionsStatus.OK) {
            const output = document.querySelector("#output");
            output.innerHTML = 
                "<div class='alert-info'>From: " +
                document.getElementById("from").value +
                ".<br />To: " +
                document.getElementById("to").value +
                ".<br /> Driving distance <i class='fas fa-road'></i> :" +
                result.routes[0].legs[0].distance.text +
                ".<br /> Duration <i class='fas fa-hourglass-start'></i> :" +
                result.routes[0].legs[0].duration.text +
                ".</div>";

                // Display route
        directionsDisplay.setDirections(result);
    } else {
        output.innerHTML = "<div class='alert-danger'>Unable to find the route. Please try again!</div>";
    }
})
}