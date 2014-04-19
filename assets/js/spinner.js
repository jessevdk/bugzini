var Spinner = function(elem) {
    this._animate_id = 0;
    this._nframes = 31;
    this._frame_period = 1 / 32;
    this._animate_start = null;
    this.elem = elem;

    return this;
}

Spinner.prototype.start = function() {
    this._animate_id = requestAnimationFrame(this.animate.bind(this));
}

Spinner.prototype.cancel = function() {
    if (this._animate_id != 0) {
        cancelAnimationFrame(this._animate_id);

        this._animate_id = 0;
        this._animate_start = null;

        this.elem.style.backgroundPositionX = '';
        this.elem.style.backgroundPositionY = '';
    }
}

Spinner.prototype.animate = function(stamp) {
    if (this._animate_start == null) {
        this._animate_start = stamp;
    }

    // In seconds
    var elapsed = (stamp - this._animate_start) / 1000.0;
    var frame = Math.floor(elapsed / this._frame_period % this._nframes);

    // Skip first frame, which is empty
    frame += 1;

    var fx = Math.floor(frame % 8);
    var fy = Math.floor(frame / 8);

    this.elem.style.backgroundPositionX = (-16 * fx) + 'px';
    this.elem.style.backgroundPositionY = (-16 * fy) + 'px';

    this.start();
}

/* vi:ts=4:et */
