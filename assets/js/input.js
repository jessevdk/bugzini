var InteractiveInput = function(input) {
    this._input = input;

    this._last_value = '';
    this._on_update_timeout = 0;
    this.update_delay = 0;

    this._input.addEventListener('keydown', this._on_keydown.bind(this));
    this._input.addEventListener('change', this._on_change.bind(this));
    this._input.addEventListener('input', this._on_change.bind(this));
    this._input.addEventListener('paste', this._on_change.bind(this));

    this.on_update = function() {}
    this.on_activate = function() {}
}

InteractiveInput.prototype._on_change = function() {
    var val = this.value();

    if (val != this._last_value) {
        this._last_value = val;
        this._emit_update(false);
    }
}

InteractiveInput.prototype._on_keydown = function(e) {
    if (e.which == 27) {
        this.cancel();
    } else if (e.which == 13) {
        this.on_activate();
    } else {
        this._on_change();
    }
}

InteractiveInput.prototype.value = function(val) {
    if (typeof val !== 'undefined') {
        this._input.value = val;
        this._emit_update();
    } else {
        return this._input.value;
    }
}

InteractiveInput.prototype.cancel = function() {
    this.value('');
    this._input.focus();
}

InteractiveInput.prototype._emit_update = function(force) {
    if (this._on_update_timeout != 0) {
        clearTimeout(this._on_update_timeout);
        this._on_update_timeout = 0;
    }

    if (force || this.update_delay == 0) {
        this.on_update(this);
    } else {
        this._on_update_timeout = setTimeout((function() {
            this._on_update_timeout = 0;
            this.on_update(this);
        }).bind(this), this.update_delay);
    }
}

/* vi:ts=4:et */