/* Tools */
function displayError(elementId, msg) {
	$("#" + elementId)
		.addClass('error')
		.after("<div class=\"error-msg box\"><div class=\"close\">x</div>" +
			msg + "</div>");
}

function close(element) {
	element.addClass('invisible');
	window.setTimeout(function() { element.hide(); element.removeClass('invisible'); }, 400);
}

function insertCloseButton(element, callback) {
	var button = $('<div class="close"><i class="icon-remove"></i></div>');
	if (callback) {
		button.click(callback);
	}

	element.prepend(button);
}

function hideAndFade(element) {
	element.addClass('invisible').delay(300).slideUp(300);
}

/* Configuration */
$(function() {
	setToggle();
	setRemoveError();
	setCloseBox();
	insertSuccessIcon();
	insertWarningIcon();
	insertErrorIcon();
	setLabelForReadOnly();
});

function setToggle() {
	$('section.toggle > header').click(function() {
		$(this).next('.body').slideToggle(500);
	});

	$('section.hover-toggle').click(function() {
		$(this).children('.body').slideToggle(500);
	});
}

function setRemoveError() {
	$('.error').focusin(function() {
		$(this).removeClass('error');
	});
}

function setCloseBox() {
	insertCloseButton($('.success-msg, .warning-msg, .error-msg, .field-error-msg'), function() {
		hideAndFade($(this).parent());
		$(this).parent()
			.prev().removeClass('error');
	});
}

function insertSuccessIcon() {
	var icon = $('<div class="msg-icon"><i class="icon-ok-sign"></i></div>');
	$(".success-msg").prepend(icon);
}

function insertWarningIcon() {
	var icon = $('<div class="msg-icon"><i class="icon-warning-sign"></i></div>');
	$(".warning-msg").prepend(icon);
}

function insertErrorIcon() {
	var icon = $('<div class="msg-icon"><i class="icon-remove-sign"></i></div>');
	$(".error-msg").prepend(icon);
}

function setLabelForReadOnly() {
	var readOnlyElements = $('*[readonly], *[disabled]');
	readOnlyElements.prev('label').addClass('label-for-readonly');
	readOnlyFixForFirefox(readOnlyElements);
}

function readOnlyFixForFirefox(elements) {
	elements.addClass('read-only');
}
