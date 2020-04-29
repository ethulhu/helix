import { elemGenerator } from './elems.js';

const _audio  = elemGenerator( 'audio' );
const _button = elemGenerator( 'button' );
const _li     = elemGenerator( 'li' );
const _ul     = elemGenerator( 'ul' );
const _video  = elemGenerator( 'video' );

export class Player {
	constructor( element ) {
		this._element = element;

		this._queue = [];
		this._current = null;
		this._tracklist = _ul();

		this._audio = _audio( { controls: true } );
		this._audio.addEventListener( 'ended', e => {
			this.playNext();
		} );

		this._video = _video( { controls: true } );
		this._video.addEventListener( 'ended', e => {
			this.playNext();
		} );

		this._element.appendChild( this._audio );
		this._element.appendChild( this._video );
		this._element.appendChild( this._tracklist );
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

		this._queue.push( item );
		this._tracklist.appendChild( _li( item.title,
			_button( '▶️', { click: () => this.skip( item ) } ),
			_button( '🚮', { click: e => { this.dequeue( item ); e.target.parentElement.parentElement.removeChild( e.target.parentElement ); } } ),
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

		const [ enabled, disabled ] =
			isAudioItem( this._current ) ?
				[ this._audio, this._video ] : [ this._video, this._audio ];
		const mimetype = this._mimetype( this._current );
		enabled.src = `/directories/${this._current.directory}/${this._current.id}?accept=${mimetype}`;
		enabled.play();
	}
}

Array.prototype.firstOrNull = function() {
	return this ? this[ 0 ] : null;
}

const isAudioItem = item => item.itemClass.startsWith( 'object.item.audioItem' );
const isVideoItem = item => item.itemClass.startsWith( 'object.item.videoItem' );
