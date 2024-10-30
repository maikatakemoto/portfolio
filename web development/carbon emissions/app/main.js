
document.addEventListener('DOMContentLoaded', function() {
    const form = document.querySelector('#calculator-form');
  
    form.addEventListener('submit', function(event) {
        event.preventDefault();
      
        // Get user inputs
        const start = document.querySelector('#start').value;
        const destination = document.querySelector('#destination').value;
        const transportationMode = document.querySelector('#transportation-mode').value;
      
        // Validate inputs
        if (!start || !destination || !transportationMode) {
            alert('Please fill in all fields.');
            return;
        }
 
 
        // If inputs are valid, proceed with calculation
        // You can implement the calculation logic here or submit the form to the server for processing
 
        // need to replace with real API call here
        const distance = simulateDistanceCalculation(start, destination);
 
 
        // calculating carbon emissions based on transportation mode?
        const emissions = calculateEmissions(transportationMode, distance);
 
 
        displayResults(emissions);
    });
 
 
    // this function below should call API to get actual distances
    // function simulateDistanceCalculation(start, destination) {
 
 
    // }
 
 
    function calculateEmissions(mode, distance) {
        // defining emission factors per km for each mode?
        const emissionFactors = {
            car: 0.21,  //kg CO2 per km
            walking: 0,
            public: 0.05 //kg CO2 per km
        };
        return (emissionFactors[mode] * distance).toFixed(2)
        // returning emissions rounded to two decimals
    }
 
 
    function displayResults(emissions) {
        alert('Your estimated carbon emissions are ${emissions} kg CO2.');
    }
 });
 