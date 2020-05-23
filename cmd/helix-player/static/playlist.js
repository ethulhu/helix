// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

import { documentFragment, elemGenerator } from './elems.js';

const _button = elemGenerator( 'table' );
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
			_button( '[ clear all ]', { click: () => this.innerHTML = '' } ),
			_table( { id: 'tracklist' } ),
		) );

		// this._updating is a flag set by `set listItems()` and unset by the mutation observer.
		this._updating = false;

		this._render();

		// TODO: maybe put observe() in connectedCallback() and disconnect() in disconnectedCallback()?
		const observer = new MutationObserver( changes => {
			if ( this._updating ) {
				this._updating = false;
				return;
			}
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

					c.removedNodes.forEach( n => {
						if ( n.tagName === 'LI' ) {
							this._sendEvent( 'trackremoved', n );
						}
					} );
				} else if ( c.target === this._current ) {
					this._sendEvent( 'currenttrackupdated', this._current );
				}
			} );
			this._render();
		} );
		observer.observe( this, { subtree: true, childList: true, attributes: true } );
	}

	// TODO: make this not rewrite the entire thing every time.
	_render() {
		const table = this.shadowRoot.getElementById( 'tracklist' );
		const df = documentFragment(
			Array.from( this.listItems ).map( ( li, i ) => _tr(
				{ part: i === ( this.currentItem - 1 ) ? 'current' : null },
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
	// - trackremoved
	_sendEvent( name, payload ) {
		this.dispatchEvent( new CustomEvent( name, { detail: payload } ) );
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
				if ( this.listItems[ newValue - 1 ].isSameNode( this._current ) ) {
					return;
				}

				const old = this._current;
				this._current = this.listItems[ newValue - 1 ];
				if ( ! this._current.isEqualNode( old ) ) {
					this._sendEvent( 'trackchanged', this._current );
				}
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
	// TODO: set fire to my own house in penance for this hack.
	set listItems( lis ) {
		this._updating = true;
		const idx = this.currentItem;
		this.innerHTML = '';
		this.appendChild( documentFragment( lis ) );
		this.currentItem = idx;
		this._render();
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
