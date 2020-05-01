import { documentFragment, elemGenerator } from './elems.js';

const _button   = elemGenerator( 'button' );
const _input    = elemGenerator( 'input' );
const _style    = elemGenerator( 'style' );

const template = documentFragment(
	_style( `
		:host {
			display: flex;
		}
		* {
			margin: 5px;
		}
		input[type=range] {
			width: 100%;
		}
	` ),
	_button( '⏯️', { id: 'playpause' } ),
	_button( '⏩', { id: 'skip' } ),
	_input( {
		id: 'slider',
		type: 'range',
		value: 0,
	} ),
);

export class HelixMediaControls extends HTMLElement {
	static get observedAttributes() {
		return [ 'duration' ];
	}

	attributeChangedCallback( attr, oldValue, newValue ) {
		switch ( attr ) {
			case 'duration':
				this._slider.max = newValue;
			case 'currentTime':
				this._slider.value = newValue;
		}
	}

	get duration() {
		return this._slider.max;
	}
	set duration( v ) {
		this.setAttribute( 'duration', v );
	}

	get currentTime() {
		return this._slider.value;
	}
	set currentTime( v ) {
		this.setAttribute( 'current-time', v );
	}

	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.appendChild( template.cloneNode( true ) );

		this._playpause.addEventListener( 'click', e => {
			this._sendEvent( 'playpause', null );
		} );
		this._skip.addEventListener( 'click', e => {
			this._sendEvent( 'skip', null );
		} );

		this._slider.addEventListener( 'input', e => {
			this._sendEvent( 'seek', e.target.value );
		} );
	}

	// events:
	// - playpause, stop,
	// - skip, back
	// - seek
	_sendEvent( name, payload ) {
		const e = new CustomEvent( name, { detail: payload } );
		this.dispatchEvent( e );
	}

	get _playpause() { return this.shadowRoot.getElementById( 'playpause' ); }
	get _skip()      { return this.shadowRoot.getElementById( 'skip'      ); }
	get _slider()    { return this.shadowRoot.getElementById( 'slider'    ); }
}

customElements.define( 'helix-media-controls', HelixMediaControls );
