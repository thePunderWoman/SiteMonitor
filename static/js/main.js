
require(
	['jquery', '/js/libs/modernizr.min.js', '/js/plugins.js', '/js/libs/mustache.js'],
	function($, m, b, must){
		$(document).on('click','.delete-site',function(e){
			e.preventDefault();
			var key = $(this).data('key');
			var curobj = $(this);
			if(confirm('Are you sure you want to delete this site from the site monitor?')) {
				$.post('/Delete', { key: key }, function(data) {
					console.log(data)
					if(data.success) {
						$(this).parent().fadeOut('slow',function(){
							$(this).parent().parent().parent().parent().remove();
						});
					}
				},'json')
			}
		});
	}
)