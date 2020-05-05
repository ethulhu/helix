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
	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.appendChild( template.cloneNode( true ) );

		this._update();
	}

	_update() {
		fetchTransports()
			.then( ts => { ts.sort( ( a, b ) => a.name.localeCompare( b ) ); return ts; } )
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
