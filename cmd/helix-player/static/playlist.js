import { documentFragment, elemGenerator } from './elems.js';

const _button = elemGenerator( 'table' );
const _style  = elemGenerator( 'style' );
const _table  = elemGenerator( 'table' );
const _td     = elemGenerator( 'td' );
const _tr     = elemGenerator( 'tr' );

// <helix-playlist>
// 	<li data-album='an album'>
// 		display text
//		<source src='…' type='…'>
// 	</li>
// </helix-playlist>

export class HelixPlaylist extends HTMLElement {

	constructor() {
		super();

		this._current = this.querySelector( 'li' );
		this.currentItem = this.currentItem;

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.appendChild( documentFragment (
			_style( `
				#tracklist tr[current] {
					color: var(--highlight-color);
				}
			` ),
			_table( { id: 'tracklist' } ),
		) );

		this.render();

		// TODO: maybe put observe() in connectedCallback() and disconnect() in disconnectedCallback()?
		const observer = new MutationObserver( changes => {
			changes.forEach( c => {
				if ( c.target === this && c.type === 'childList' ) {
					let current = this._current;
					if ( Array.from( c.removedNodes ).includes( this._current ) ) {
						current = c.nextSibling;
						while ( current && current.tagName !== 'LI' ) {
							current = current.nextElementSibling;
						}
					}
					if ( ! current && c.addedNodes ) {
						current = c.addedNodes[ 0 ];
						while ( current && current.tagName !== 'LI' ) {
							current = current.nextElementSibling;
						}
					}
					this.currentItem = this.indexOf( current ) + 1;
				} else if ( c.target === this._current ) {
					this._sendEvent( 'currenttrackupdated' );
				}
			} );
			this.render();
		} );
		observer.observe( this, { subtree: true, childList: true, attributes: true } );
	}

	// TODO: make this not rewrite the entire thing every time.
	render() {
		const table = this.shadowRoot.getElementById( 'tracklist' );
		const df = documentFragment(
			Array.from( this.listItems ).map( ( li, i ) => _tr(
				{ current: i === ( this.currentItem - 1 ) },
				_td( li.textContent ),
				_td( _button( '▶️', { click: () => this.currentItem = ( i + 1 ) } ) ),
				_td( _button( '🚮', { click: () => this.removeChild( this.listItems[ i ] ) } ) ),
			) ),
		);
		table.innerHTML = '';
		table.appendChild( df );
	}

	// events:
	// - trackchanged
	// - currenttrackupdated
	_sendEvent( name ) {
		this.dispatchEvent( new CustomEvent( name, { detail: this._current } ) );
	}

	static get observedAttributes() {
		return [ 'current-item' ];
	}
	attributeChangedCallback( name, oldValue, newValue ) {
		switch ( name ) {
			case 'current-item':
				if ( ! ( 1 <= newValue && newValue <= this.listItems.length ) ) {
					return;
				}
				if ( this.listItems[ newValue - 1 ] === this._current ) {
					return;
				}

				if ( this._current ) {
					this._current.removeAttribute( 'current' );
				}
				this._current = this.listItems[ newValue - 1 ];
				this._current.setAttribute( 'current', '' );
				this._sendEvent( 'trackchanged' );
		}
	}

	indexOf( node ) {
		return Array.from( this.children )
			.filter( el => el.tagName === 'LI' )
			.indexOf( node );
	}

	get currentItem() {
		return this.indexOf( this._current ) + 1;
	}
	set currentItem( value ) {
		this.setAttribute( 'current-item', value );
	}

	get listItems() {
		return this.getElementsByTagName( 'li' );
	}

	skip() {
		this.currentItem++;
	}
	back() {
		if ( this.currentItem > 0 ) {
			this.currentItem--;
		}
	}
}

customElements.define( 'helix-playlist', HelixPlaylist );
