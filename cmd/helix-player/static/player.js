import { elemGenerator } from './elems.js';

const _audio  = elemGenerator( 'audio' );
const _button = elemGenerator( 'button' );
const _details  = elemGenerator( 'details' );
const _li     = elemGenerator( 'li' );
const _summary  = elemGenerator( 'summary' );
const _ul     = elemGenerator( 'ul' );
const _video  = elemGenerator( 'video' );

export class Player {
	constructor( element ) {
		this._element = element;

		this._queue = [];
		this._current = null;
		this._tracklist = _ul();

		// Incrementing this provides a source of unique IDs.
		this._playlistIds = 0;

		this._audio = _audio( { controls: true } );
		this._audio.addEventListener( 'ended', e => {
			this.playNext();
		} );

		this._video = _video( { controls: true, style: 'display: none;' } );
		this._video.addEventListener( 'ended', e => {
			this.playNext();
		} );


		this._element.appendChild( this._audio );
		this._element.appendChild( this._video );
		this._element.appendChild( _details( _summary( 'playlist' ), this._tracklist ) );
	}

	_newPlaylistId() {
		return this._playlistIds++;
	}

	_mimetype( item ) {
		// TODO: reorder item.mimetypes into [ canPlayType == 'probably' ] + [ canPlayType == 'maybe' ].
		return isAudioItem( item ) ? item.mimetypes.filter( m => this._audio.canPlayType( m ) ).firstOrNull() :
			isVideoItem( item ) ? item.mimetypes.filter( m => this._video.canPlayType( m ) ).firstOrNull() :
			null;
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
			this.playNext();
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

	playNext() {
		if ( ! this._queue ) {
			// empty playlist.
			return;
		}

		if ( ! this._current ) {
			// start playing.
			this._current = this._queue[ 0 ];
			this._element.querySelector( `li[data-playlist-id='${this._current.playlistId}']` )
				.classList.add( 'playing' );
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

		this._element.querySelector( `li[data-playlist-id='${this._current.playlistId}']` )
			.classList.remove( 'playing' );
		this._current = this._queue[ index + 1 ];
		this.play();
		this._element.querySelector( `li[data-playlist-id='${this._current.playlistId}']` )
			.classList.add( 'playing' );
	}

	play() {
		if ( ! this._current ) {
			return;
		}

		const [ enabled, disabled ] =
			isAudioItem( this._current ) ?
				[ this._audio, this._video ] : [ this._video, this._audio ];

		disabled.style.display = 'none';

		const mimetype = this._mimetype( this._current );
		enabled.src = `/directories/${this._current.directory}/${this._current.id}?accept=${mimetype}`;
		enabled.style.display = 'block';
		enabled.play();
	}
}

Array.prototype.firstOrNull = function() {
	return this ? this[ 0 ] : null;
}

const isAudioItem = item => item.itemClass.startsWith( 'object.item.audioItem' );
const isVideoItem = item => item.itemClass.startsWith( 'object.item.videoItem' );
