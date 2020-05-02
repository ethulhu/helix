async function getJSON( url ) {
	const rsp = await fetch( url, {
		headers: { Accept: 'application/json' },
	} );
	return rsp.json();
}

export async function fetchDirectories() {
	return getJSON( '/directories/' );
}

export async function fetchObject( directory, id ) {
	return getJSON( `/directories/${directory}/${id}` );
}

export const rootObject = '0';
