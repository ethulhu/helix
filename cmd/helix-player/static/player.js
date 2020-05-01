import { elemGenerator } from './elems.js';

const _audio   = elemGenerator( 'audio' );
const _button  = elemGenerator( 'button' );
const _details = elemGenerator( 'details' );
const _div     = elemGenerator( 'div' );
const _li      = elemGenerator( 'li' );
const _summary = elemGenerator( 'summary' );
const _ul      = elemGenerator( 'ul' );
const _video   = elemGenerator( 'video' );

export class Player {

	get currentTime() { return this._playingElement.currentTime; }
	set currentTime( v ) { this._playingElement.currentTime = v; }

	get duration() { return this._playingElement.duration; }

	constructor( element ) {
		this._element = element;

		this._queue = [];
		this._current = null;
		this._tracklist = _ul();

		// Incrementing this provides a source of unique IDs.
		this._playlistIds = 0;


		this._audio = _audio( {
			ended: () => this.skip(),
			timeupdate: () => this._sendEvent( 'timeupdate', null ),
			durationchange: () => this._sendEvent( 'durationchange', null ),
		} );

		this._video = _video( {
			controls: true,
			style: 'display: none;',
			ended: () => this.skip(),
		} );


		this._element.appendChild( this._audio );
		this._element.appendChild( this._video );
		this._element.appendChild( _details( _summary( 'playlist' ), this._tracklist ) );
	}

	_sendEvent( name, payload ) {
		const e = new CustomEvent( name, { detail: payload } );
		this._element.dispatchEvent( e );
	}

	get _playingElement() {
		return this._current && isVideoItem( this._current ) ? this._video : this._audio;
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
		this._playingElement.pause();

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

	playPause() {
		if ( this._playingElement.paused ) {
			this.play();
		} else {
			this._playingElement.pause();
		}
	}

	play() {
		if ( ! this._current ) {
			return;
		}

		const [ enabled, disabled ] =
			isAudioItem( this._current ) ?
				[ this._audio, this._video ] : [ this._video, this._audio ];

		const mimetype = this._mimetype( this._current );
		const url = `/directories/${this._current.directory}/${this._current.id}?accept=${mimetype}`;
		if ( enabled.src.endsWith( url ) ) {
			enabled.play();
			return;
		}
		enabled.src = url;

		disabled.style.display = 'none';
		this._element.querySelectorAll( 'li.playing' ).forEach( el => el.classList.remove( 'playing' ) );

		enabled.style.display = 'block';
		this._element.querySelector( `li[data-playlist-id='${this._current.playlistId}']` )
			.classList.add( 'playing' );

		enabled.play();
	}
}

Array.prototype.firstOrNull = function() {
	return this ? this[ 0 ] : null;
}

const isAudioItem = item => item.itemClass.startsWith( 'object.item.audioItem' );
const isVideoItem = item => item.itemClass.startsWith( 'object.item.videoItem' );
