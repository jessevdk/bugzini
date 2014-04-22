var Popover = function(elem, loc) {
	var overlay = document.createElement('div');
	overlay.classList.add('popover-overlay');

	var body = document.body;
	body.appendChild(overlay);

	var popover = document.createElement('div');
	popover.classList.add('popover');

	if (elem) {
		popover.appendChild(elem);
	}

	popover.style.left = (loc[0] - 24) + 'px';
	popover.style.top = (loc[1] + 24) + 'px';

	body.appendChild(popover);

	this._overlay = overlay;
	this._popover = popover;

	this._overlay.setAttribute('tabindex', '0');
	this._overlay.addEventListener('keydown', this._on_keydown.bind(this));
	this._overlay.addEventListener('mousedown', this.close.bind(this));

	this._popover.setAttribute('tabindex', '0');
	this._popover.addEventListener('keydown', this._on_keydown.bind(this));

	// Force layout
	overlay.focus();

	// Animate
	overlay.classList.add('shown');

	body.classList.add('modal');
	return this;
}

Popover.prototype._on_keydown = function(e) {
	if (e.which == 27) {
		this.close();
	}
}

Popover.prototype.close = function() {
	if (!this._overlay) {
		return;
	}

	var body = document.body;

	body.classList.remove('modal');
	
	body.removeChild(this._overlay);
	body.removeChild(this._popover);

	this._overlay = null;
	this._popover = null;
}