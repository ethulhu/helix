import { elemGenerator } from './elems.js';

const _button  = elemGenerator( 'button' );
const _details = elemGenerator( 'details' );
const _div     = elemGenerator( 'div' );
const _li      = elemGenerator( 'li' );
const _summary = elemGenerator( 'summary' );
const _ul      = elemGenerator( 'ul' );
const _video   = elemGenerator( 'video' );

export class Player {

	constructor( element, player ) {
		this._element = element;
		this._player = player;

		this._queue = [];
		this._current = null;
		this._tracklist = _ul();

		// Incrementing this provides a source of unique IDs.
		this._playlistIds = 0;


		this._element.appendChild( _details( _summary( 'playlist' ), this._tracklist ) );
	}

	_sendEvent( name, payload ) {
		const e = new CustomEvent( name, { detail: payload } );
		this._element.dispatchEvent( e );
	}

	_newPlaylistId() {
		return this._playlistIds++;
	}

	_mimetype( item ) {
		return item.mimetypes.filter( m => this._player.canPlayType( m ) ).firstOrNull();
	}

	canPlay( item ) {
		return !! this._mimetype( item );
	}

	enqueue( item ) {
		if ( ! this.canPlay( item ) ) {
			throw `cannot enqueue item: directory ${item.directory}, id ${item.id}, class ${item.itemClass}`;
		}

		// Clone the item.
		let playlistItem = {};
		Object.assign( playlistItem, item );
		playlistItem.playlistId = this._newPlaylistId();

		this._queue.push( playlistItem );
		this._tracklist.appendChild( _li( playlistItem.title,
			{ 'data-playlist-id': playlistItem.playlistId },
			_button( '▶️', { click: () => this.skip( playlistItem ) } ),
			_button( '🚮', { click: e => {
				this.dequeue( playlistItem );
				e.target.parentElement.parentElement.removeChild( e.target.parentElement );
			} } ),
		) );

		if ( ! this._current ) {
			this.skip();
		}
	}

	dequeue( item ) {
		if ( item === this._current ) {
			// TODO: throw?
			return;
		}
		this._queue = this._queue.filter( i => i !== item );
	}

	skip( item ) {
		this._current = item;
		this.play();
	}

	skip() {
		if ( ! this._queue ) {
			// empty playlist.
			return;
		}

		if ( ! this._current ) {
			// start playing.
			this._current = this._queue[ 0 ];
			this.play();
			return;
		}

		const index = this._queue.indexOf( this._current );

		if ( index < 0 ) {
			throw `item not in playlist`;
		}

		if ( index === ( this._queue.length - 1 ) ) {
			// playlist ended.
			return;
		}

		this._current = this._queue[ index + 1 ];
		this.play();
	}

	play() {
		if ( ! this._current ) {
			return;
		}

		const mimetype = this._mimetype( this._current );
		const url = `/directories/${this._current.directory}/${this._current.id}?accept=${mimetype}`;

		if ( this._player.src.endsWith( url ) ) {
			this._player.play();
			return;
		}
		this._player.src = url;

		this._element.querySelectorAll( 'li.playing' ).forEach( el => el.classList.remove( 'playing' ) );

		this._element.querySelector( `li[data-playlist-id='${this._current.playlistId}']` )
			.classList.add( 'playing' );

		this._player.play();
	}
}

Array.prototype.firstOrNull = function() {
	return this ? this[ 0 ] : null;
}
