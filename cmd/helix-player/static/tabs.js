// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

import { documentFragment, elemGenerator } from './elems.js';

const _button = elemGenerator( 'button' );

// <helix-tabs tab='tab-13'>
// 	<helix-tab id='tab-12' title='tab 1?'>
// 		<p>hello I am content</p>
// 	</helix-tab>
// 	<helix-tab id='tab-13' title='tab 2?'>
// 		<p>bride of content</p>
// 	</helix-tab>
// </helix-tabs>
// 
// OR
// 
// <helix-tabs>
// 	<helix-tab id='tab-12' title='tab 1?' active>
// 		<p>hello I am content</p>
// 	</helix-tab>
// 	<helix-tab id='tab-13' title='tab 2?'>
// 		<p>bride of content</p>
// 	</helix-tab>
// </helix-tabs>

export class HelixTab extends HTMLElement {
	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.innerHTML = `
			<slot></slot>
		`;
	}

	static get observedAttributes() {
		return [ 'title', 'active' ];
	}

	get title() {
		return this.getAttribute( 'title' );
	}
	set title( value ) {
		this.setAttribute( 'title', value );
	}

	get active() {
		const value = this.getAttribute( 'active' )
		return ( !! value ) || value === '';
	}
	set active( value ) {
		if ( value && ! this.active ) {
			this.setAttribute( 'active', '' );
		} else if ( ! value && this.active ) {
			this.removeAttribute( 'active' );
		}
	}
}

export class HelixTabs extends HTMLElement {
	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.innerHTML = `
			<style>
				::slotted(helix-tab) {
					display: none;
				}
				::slotted(helix-tab[active]) {
					display: block;
				}
			</style>
			<div id='buttons'></div>
			<slot></slot>
		`;

		const observer = new MutationObserver( changes => {
			const isChildrenChanged = c => c.target === this && c.type === 'childList';
			const isTabAttributeChanged =
				c => this.tabs.some( t => t.isSameNode( c.target ) )
					&& c.type === 'attributes'
					&& [ 'active', 'title' ].includes( c.attributeName );
			const rerender = changes.some( c => isChildrenChanged( c ) || isTabAttributeChanged( c ) );
			if ( rerender ) {
				this._render();
			}
		} );
		observer.observe( this, { subtree: true, childList: true, attributes: true } );

		this._render();
	}

	_sendEvent( name, payload ) {
		this.dispatchEvent( new CustomEvent( name, { detail: payload } ) );
	}

	_render() {
		const df = documentFragment(
			this.tabs.map( tab => _button(
				tab.title,
				{
					part: tab.id === this.tab ? 'tab active' : 'tab',
					click: () => this.tabs.forEach( t => {
						t.active = t.isSameNode( tab );
					} ),
				}
			) ),
		)
		this._buttons.innerHTML = '';
		this._buttons.appendChild( df );
		this.tab = this.tab;
	}

	static get observedAttributes() {
		return [ 'tab' ];
	}
	attributeChangedCallback( name, oldValue, newValue ) {
		switch ( name ) {
			case 'tab':
				const newActive = this.querySelector( `#${newValue}` );
				if ( ! ( newActive && newActive.tagName === 'HELIX-TAB' ) ) {
					this.tab = oldValue;
					return;
				}

				// this blocks on <helix-tab> being defined
				// because otherwise it sets the .active value on the object
				// before the setter exists.
				customElements.whenDefined( 'helix-tab' ).then( () => {
					this.tabs.forEach( t => { t.active = t.isSameNode( newActive ) } )
					this._sendEvent( 'tabchanged', null );
				} );
		}
	}

	get tab() {
		const active = this.tabs.filter( tab => tab.active );
		if ( active.length > 0 ) {
			return active[ 0 ].id;
		}
		if ( this.tabs.length > 0 ) {
			return this.tabs[ 0 ].id;
		}
		return null;
	}
	set tab( id ) {
		this.setAttribute( 'tab', id );
	}

	get tabs() {
		return Array.from( this.getElementsByTagName( 'helix-tab' ) );
	}
	get _buttons() {
		return this.shadowRoot.getElementById( 'buttons' );
	}
}

customElements.define( 'helix-tab', HelixTab );
customElements.define( 'helix-tabs', HelixTabs );
