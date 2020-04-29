import { elemGenerator } from './elems.js';

const _button = elemGenerator( 'button' );
const _input = elemGenerator( 'input' );
const _label = elemGenerator( 'label' );
const _li = elemGenerator( 'li' );
const _ul = elemGenerator( 'ul' );

export class Directory {
	constructor( element, api, player ) {
		this._element = element;
		this._api = api;
		this._player = player;
	}

	loadDirectories() {
		this._api.directories()
			.then( ds => _ul( ds.map( this.newDirectory.bind( this ) ) ) )
			.then( ul => {
				this._element.innerHTML = '';
				this._element.appendChild( ul );
			} );
	}

	newDirectory( d ) {
		return _li(
			_input( {
				'data-directory': d.udn,
				id: d.udn,
				type: 'checkbox',
				change: e => {
					if ( e.target.checked ) {
						this._api.object( d.udn, this._api.rootObject )
							.then( o => e.target.parentElement.appendChild(
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
				'data-id': o.id,
				'data-directory': o.directory,
				id: `${o.directory}/${o.id}`,
				type: 'checkbox',
				change: e => {
					if ( e.target.checked ) {
						this._api.object( o.directory, o.id )
							.then( o => e.target.parentElement.appendChild(
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
				'data-id': o.id,
				'data-directory': o.directory,
				'click': e => this._player.enqueue( o ),
				disabled: ! this._player.canPlay( o ),
			} )
		);
	}
}
