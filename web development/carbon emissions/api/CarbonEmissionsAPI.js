const apiKey = 'Nmg24Rr0pZ6Es2Zx5FYzJA'

// Default car: 2021 Toyota Corolla
// id: 329f0cf0-6777-48f0-9dce-ef1d6fcfc4d6

// Default bus: Ford Transit T150 Wagon FFV
// id: ad87879b-3d40-45a6-8e4f-bcf28bd49ae0

// Function to calculate distance using Google Maps Distance Matrix API
function calculateDistance(origin, destination) {
    var service = new google.maps.DistanceMatrixService();
    service.getDistanceMatrix(
        {
            origins: [origin],
            destinations: [destination],
            travelMode: 'DRIVING',
            unitSystem: google.maps.UnitSystem.IMPERIAL,
        },
        async function (response, status) {
            if (status === 'OK') {
                var distance = response.rows[0].elements[0].distance.value; // Distance in meters
                var distanceKm = distance / 1609.34; // Convert distance to miles

                try {
                    const buttonPressed = "car"; 
                    let vehicleModelId;
                    if (buttonPressed) {
                        vehicleModelId = "329f0cf0-6777-48f0-9dce-ef1d6fcfc4d6"; 
                    } else if (buttonPressed === "bus") {
                        vehicleModelId = "329f0cf0-6777-48f0-9dce-ef1d6fcfc4d6"; 
                    }

                    // Call function to estimate carbon emissions
                    estimateCarbonEmissions(vehicleModelId, 'mi', distanceMiles);
                } catch (error) {
                    console.error('Error:', error.message);
                }
            } else {
                console.error('Error with Distance Matrix service:', status);
            }
        }
    );
}

// Function to estimate carbon emissions using Carbon Interface API
async function estimateCarbonEmissions(vehicleModelId, distanceUnit, distanceValue) {
    const apiUrl = 'https://www.carboninterface.com/api/v1/estimates';

    const payload = {
        type: 'vehicle',
        distance_unit: distanceUnit,
        distance_value: distanceValue,
        vehicle_model_id: vehicleModelId, 
    };

    try {
        const response = await fetch(apiUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${apiKey}`,
            },
            body: JSON.stringify(payload),
        });

        if (!response.ok) {
            throw new Error('Failed to fetch carbon emissions estimate');
        }

        const data = await response.json();
        console.log('Carbon emissions estimate:', data);

        // Process the response (e.g., display carbon emissions)
        const carbonEmissions = data?.data?.attributes?.carbon_kg; // Carbon emissions in kilograms
        console.log(`Estimated carbon emissions: ${carbonEmissions} kg`);

        // Display carbon emissions or perform further actions
        displayCarbonEmissions(carbonEmissions);
    } catch (error) {
        console.error('Error fetching carbon emissions estimate:', error.message);
    }
}

// Function to display carbon emissions (example implementation)
function displayCarbonEmissions(carbonEmissions) {
    // Example: Display carbon emissions in the UI
    document.getElementById('carbonEmissions').textContent = `Estimated carbon emissions: ${carbonEmissions} kg`;
}

// Event listener for form submission (calculate button)
document.querySelector('.form-horizontal').addEventListener('submit', function (event) {
    event.preventDefault(); // Prevent default form submission

    // Get origin and destination input values
    const origin = document.getElementById('from').value;
    const destination = document.getElementById('to').value;

    // Calculate distance between origin and destination
    calculateDistance(origin, destination);
});