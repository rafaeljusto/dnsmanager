$(function() {
	$("#new").click(function() {
		window.location.href="/domain";
	});

	$("#help").click(function() {
		$("#help-box").dialog({
			height: 500,
			width: 680,
			title: "Help",
			buttons: {
				"OK": function() {
					$(this).dialog("close");
				}
			}
		}); 
	});
})