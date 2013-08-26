(function(sawsij, $, undefined) {

	$(function() {
		wireClickRows();
    });


	function wireClickRows(){
		$("table.table-clickrows").each(function(){
			$(this).find("tr").each(function(){
				var link = $(this).find('a').first();
				if(typeof link.attr('href') != "undefined"){
					link.click(function(e){
						e.preventDefault();
					});					
				}
				$(this).click(function(){
					window.location = link.attr('href');
				});
			});
		});
	}


	$('.datepicker').datepicker();

}(window.sawsij = window.sawsij || {}, jQuery));