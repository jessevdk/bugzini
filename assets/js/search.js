var Search = function(elem) {
    this.elem = elem;

    var e = new $(elem);

    var input = e.query('input');
    this.input = new InteractiveInput(input);

    this._close = e.query('.close-icon');
    this._params = e.query('.params-icon');

    this._all = [input, elem, this._close, this._params];

    input.addEventListener('focus', this._on_focus.bind(this));
    input.addEventListener('blur', this._on_blur.bind(this));
    
    this._close.addEventListener('mousedown', this._on_close.bind(this));
    this._params.addEventListener('mousedown', this._on_params.bind(this));

    this.on_params = function() {}
}

Search.prototype.value = function() {
    return this.input.value();
}

Search.prototype._on_params = function(e) {
    this.on_params(e);
}

Search.prototype._on_focus = function() {
    this._all.each(function (item) {
        item.classList.add('focus');
    });
}

Search.prototype._on_blur = function() {
    this._all.each(function (item) {
        item.classList.remove('focus');
    });
}

Search.prototype._on_close = function() {
    this.input.cancel();
}

/* vi:ts=4:et */
