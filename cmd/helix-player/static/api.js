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
