import { elemGenerator } from './elems.js';

const _button   = elemGenerator( 'button' );
const _input    = elemGenerator( 'input' );
const _label    = elemGenerator( 'label' );
const _li       = elemGenerator( 'li' );
const _style    = elemGenerator( 'style' );
const _template = elemGenerator( 'template' );
const _ul       = elemGenerator( 'ul' );

async function _fetchDirectories() {
	const rsp = await fetch( '/directories/', {
		headers: { Accept: 'application/json' },
	} );
	return rsp.json()
}
async function _fetchObject( directory, id ) {
	const rsp = await fetch( `/directories/${directory}/${id}`, {
		headers: { Accept: 'application/json' },
	} );
	return rsp.json()
}
const rootObject = '0';

const template = _template(
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
		this.shadowRoot.appendChild( template.content.cloneNode( true ) );

		_fetchDirectories()
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
						_fetchObject( d.udn, rootObject )
							.then( o => target.parentElement.appendChild(
								o.itemClass.startsWith( 'object.container' ) ?
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
		return o.itemClass.startsWith( 'object.container' ) ? this.newContainer( o ) : this.newItem( o );
	}

	newContainer( o ) {
		return _li(
			_input( {
				id: `${o.directory}/${o.id}`,
				type: 'checkbox',
				change: e => {
					if ( e.target.checked ) {
						const target = e.target;
						_fetchObject( o.directory, o.id )
							.then( o => target.parentElement.appendChild(
								_ul( o.children.map( this.newObject.bind( this ) ) ) )
							)
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
				'click': e => this._sendEvent( 'enqueue', o ),
				// TODO: figure out what to do with this.
				// While there are things a browser cannot play,
				// a transport might be able to,
				// and this is supposed to be reusable.
				// disabled: ! this._player.canPlay( o ),
			} )
		);
	}
}

customElements.define( 'helix-directory-tree', HelixDirectoryTree );
