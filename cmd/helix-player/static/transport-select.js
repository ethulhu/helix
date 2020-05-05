import { documentFragment, elemGenerator } from './elems.js';
import { fetchControlPoint, fetchTransports } from './api.js';

const _option = elemGenerator( 'option' );
const _select = elemGenerator( 'select' );

const template = documentFragment(
	_select(
		{ id: 'select' },
		_option( 'no transport', { value: 'none' } ),
		_option( '────────────', { disabled: true } ),
	),
);

export class HelixTransportSelect extends HTMLElement {
	static get observedAttributes() {
		return [ 'value' ];
	}

	attributeChangedCallback( name, oldValue, newValue ) {
		switch ( name ) {
			case 'value':
				this._select.value = newValue;
		}
	}

	get value() {
		return this._select.value;
	}
	set value( value ) {
		this.setAttribute( 'value', value );
	}

	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.appendChild( template.cloneNode( true ) );

		this._select.addEventListener( 'change', e => {
			this._sendEvent( 'change', null );
		} );

		this._update();
	}

	_sendEvent( name, payload ) {
		this.dispatchEvent( new CustomEvent( name, { detail: payload } ) );
	}

	_update() {
		fetchTransports()
			.then( ts => { ts.sort( ( a, b ) => a.name.localeCompare( b.name ) ); return ts; } )
			.then( ts => ts.forEach( 
				t => this._select.appendChild( _option( t.name, { value: t.id } ) )
			) )
			.then( fetchControlPoint )
			.then( cp => { this._select.value = cp.transport; } )
			.catch( console.error );
	}

	get _select() {
		return this.shadowRoot.getElementById( 'select' );
	}
}

customElements.define( 'helix-transport-select', HelixTransportSelect );
