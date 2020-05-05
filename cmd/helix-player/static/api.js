async function getJSON( url ) {
	const rsp = await fetch( url, {
		headers: { Accept: 'application/json' },
	} );
	return rsp.json();
}

async function postForm( url, data ) {
	let fd = new FormData();
	Object.entries( data ).forEach( ( [ k, v ] ) => {
		fd.append( k, v );
	} );
	return fetch( url, {
		method: 'POST',
		body: fd,
	} );
}

// ContentDirectory APIs.

export async function fetchDirectories() {
	return getJSON( '/directories/' );
}

export async function fetchObject( directory, id ) {
	return getJSON( `/directories/${directory}/${id}` );
}

export const rootObject = '0';


// AVTransport APIs.

export async function fetchTransports() {
	return getJSON( '/transports/' );
}

export async function fetchTransport( udn ) {
	return getJSON( `/transports/${udn}` );
}

export async function playTransport( udn ) {
	return postForm( `/transports/${udn}`, { action: 'play' } );
}

export async function pauseTransport( udn ) {
	return postForm( `/transports/${udn}`, { action: 'pause' } );
}

export async function stopTransport( udn ) {
	return postForm( `/transports/${udn}`, { action: 'stop' } );
}

// Control Point APIs.

export async function fetchControlPoint() {
	return getJSON( `/control-point/` );
}
export async function playControlPoint() {
	return postForm( `/control-point/`, { state: 'playing' } );
}
export async function pauseControlPoint() {
	return postForm( `/control-point/`, { state: 'paused' } );
}
export async function stopControlPoint() {
	return postForm( `/control-point/`, { state: 'stopped' } );
}
export async function setControlPointTransport( udn ) {
	return postForm( `/control-point/`, { transport: udn } );
}

// Control Point APIs.

export async function fetchQueue() {
	return getJSON( `/queue/` );
}
export async function appendToQueue( directory, id ) {
	return postForm( `/queue/`, { directory: directory, object: id } )
		.then( rsp => rsp.json() );
}
export async function setCurrentQueueTrack( id ) {
	return postForm( `/queue/`, { current: id } );
}
export async function removeAllFromQueue() {
	return postForm( `/queue/`, { remove: 'all' } );
}
export async function removeTrackFromQueue( id ) {
	return postForm( `/queue/`, { remove: id } );
}
