
require(
	['jquery', '/js/libs/modernizr.min.js', '/js/plugins.js', '/js/libs/mustache.js'],
	function($, m, b, must){
		$(document).on('click','.delete-site',function(e){
			e.preventDefault();
			var key = $(this).data('key');
			var curobj = $(this);
			if(confirm('Are you sure you want to delete this site from the site monitor?')) {
				$.post('/Delete', { key: key }, function(data) {
					//console.log(data)
					if(data.success) {
						var row = $(curobj).parent().parent().parent().parent().parent();
						$(row).fadeOut('fast',function(){
							$(row).remove();
						});
					}
				},'json')
			}
		});
		$(document).on('click','.delete-notifier',function(e){
			e.preventDefault();
			var key = $(this).data('key');
			var parent = $(this).data('parent');
			var curobj = $(this);
			if(confirm('Are you sure you want to delete this email from the site notification list?')) {
				$.post('/DeleteNotifier', { key: key, parent: parent }, function(data) {
					if(data.success) {
						var row = $(curobj).parent().parent().parent().parent().parent();
						$(row).fadeOut('fast',function(){
							$(row).remove();
						});
					}
				},'json')
			}
		});
		$(document).on('click','.statusgroup p', function(e){
			e.preventDefault();
			if($(this).parent().hasClass('up')) {
				$(this).parent().removeClass('up').addClass('down');
			} else {
				$(this).parent().removeClass('down').addClass('up');
			}
		});
	}
)