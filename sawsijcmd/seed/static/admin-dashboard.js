google.load("visualization", "1", {packages:["corechart"]});
google.setOnLoadCallback(drawCharts);

function drawCharts() {
	var piedata = google.visualization.arrayToDataTable([
	  ['Pie', 'Amount'],
	  ['Eaten', 30],
	  ['Not Eaten', 80],	  
	]);

	var options = {	  
	  legend: 'none',
      pieSliceText: 'label',	
      chartArea:{width:"100%",height:"95%"}  
	};

	var piechart = new google.visualization.PieChart(document.getElementById('piechart'));
	piechart.draw(piedata, options);
}