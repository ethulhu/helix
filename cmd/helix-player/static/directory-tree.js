// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

import { documentFragment, elemGenerator } from './elems.js';
import { fetchDirectories, fetchObject, rootObject } from './api.js';

const _button = elemGenerator( 'button' );
const _input  = elemGenerator( 'input' );
const _label  = elemGenerator( 'label' );
const _li     = elemGenerator( 'li' );
const _style  = elemGenerator( 'style' );
const _ul     = elemGenerator( 'ul' );

const template = documentFragment(
	_style( `
		button:disabled {
			text-decoration: line-through;
		}
		input[type=checkbox] ~ ul {
			visibility: hidden;
		}
		input[type=checkbox]:checked ~ ul {
			visibility: visible;
		}
	` ),
);

export class HelixDirectoryTree extends HTMLElement {
	constructor() {
		super();

		this.attachShadow( { mode: 'open' } );
		this.shadowRoot.appendChild( template.cloneNode( true ) );

		fetchDirectories()
			.then( ds => { ds.sort( ( a, b ) => a.name.localeCompare( b.name ) ); return ds; } )
			.then( ds => _ul( ds.map( this.newDirectory.bind( this ) ) ) )
			.then( ul => this.shadowRoot.appendChild( ul ) );
	}

	_sendEvent( name, payload ) {
		const e = new CustomEvent( name, { detail: payload } );
		this.dispatchEvent( e );
	}

	newDirectory( d ) {
		return _li(
			_input( {
				id: d.udn,
				type: 'checkbox',
				change: e => {
					if ( e.target.checked ) {
						const target = e.target;
						fetchObject( d.udn, rootObject )
							.then( o => target.parentElement.appendChild(
								isContainer( o ) ?
									_ul( o.children.map( this.newObject.bind( this ) ) ) :
									_ul( this.newObject( o ) ) )
							)
							.catch( console.error );
					} else {
						const ul = e.target.parentElement.querySelector( 'ul' );
						e.target.parentElement.removeChild( ul );
					}
				},
			} ),
			_label( d.name, { for: d.udn } ),
		);
	}

	newObject( o ) {
		return isContainer( o ) ? this.newContainer( o ) : this.newItem( o );
	}

	newContainer( o ) {
		return _li(
			_input( {
				id: `${o.directory}/${o.id}`,
				type: 'checkbox',
				change: e => {
					if ( e.target.checked ) {
						const target = e.target;
						fetchObject( o.directory, o.id )
							.then( o => target.parentElement.appendChild(
								_ul(
									o.children.some( isItem ) ? this.addAll( o.children ) : '',
									o.children.map( this.newObject.bind( this ) ),
								),
							) )
							.catch( console.error );
					} else {
						const ul = e.target.parentElement.querySelector( 'ul' );
						e.target.parentElement.removeChild( ul );
					}
				},
			} ),
			_label( o.title, { for: `${o.directory}/${o.id}` } ),
		);
	}

	newItem( o ) {
		return _li(
			_button( o.title, {
				'click': () => this._sendEvent( 'enqueue', o ),
				// TODO: figure out what to do with this.
				// While there are things a browser cannot play,
				// a transport might be able to,
				// and this is supposed to be reusable.
				// disabled: ! this._player.canPlay( o ),
			} ),
		);
	}

	addAll( os ) {
		return _li(
			_button( '[ add all ]', {
				'click': () => {
					os.filter( isItem ).forEach( o => this._sendEvent( 'enqueue', o ) );
				},
			} ),
		);
	}
}

const isContainer = o => o.itemClass.startsWith( 'object.container' );
const isItem      = o => o.itemClass.startsWith( 'object.item' );

customElements.define( 'helix-directory-tree', HelixDirectoryTree );
