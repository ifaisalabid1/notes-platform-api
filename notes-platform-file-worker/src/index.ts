export interface Env {
	NOTES_BUCKET: R2Bucket;
	API_BASE_URL: string;
	WORKER_API_SECRET: string;
}

type FileMetadata = {
	note_id: string;
	title: string;
	original_file_name: string;
	stored_object_key: string;
	file_content_type: string;
	file_size_bytes: number;
};

export default {
	async fetch(request: Request, env: Env): Promise<Response> {
		try {
			return await handleRequest(request, env);
		} catch (error) {
			console.error(error);
			return jsonError(500, "internal_error", "Something went wrong.");
		}
	},
};

async function handleRequest(request: Request, env: Env): Promise<Response> {
	if (request.method !== "GET" && request.method !== "HEAD") {
		return jsonError(405, "method_not_allowed", "Method not allowed.", {
			Allow: "GET, HEAD",
		});
	}

	const url = new URL(request.url);
	const match = url.pathname.match(/^\/notes\/([0-9a-fA-F-]{36})$/);

	if (!match) {
		return jsonError(404, "not_found", "File was not found.");
	}

	const noteID = match[1];

	const metadata = await fetchFileMetadata(env, noteID);

	if (!metadata) {
		return jsonError(404, "not_found", "File was not found.");
	}

	const rangeHeader = request.headers.get("Range");
	const object = await getR2Object(env, metadata.stored_object_key, rangeHeader);

	if (!object) {
		return jsonError(404, "not_found", "File was not found.");
	}

	const headers = buildFileHeaders(metadata, object);

	if (request.method === "HEAD") {
		return new Response(null, {
			status: rangeHeader ? 206 : 200,
			headers,
		});
	}

	return new Response(object.body, {
		status: rangeHeader ? 206 : 200,
		headers,
	});
}

async function fetchFileMetadata(env: Env, noteID: string): Promise<FileMetadata | null> {
	const apiBaseURL = env.API_BASE_URL.replace(/\/+$/, "");
	const response = await fetch(`${apiBaseURL}/internal/notes/${noteID}/file`, {
		method: "GET",
		headers: {
			"X-Worker-Secret": env.WORKER_API_SECRET,
			Accept: "application/json",
		},
	});

	if (response.status === 404) {
		return null;
	}

	if (!response.ok) {
		throw new Error(`metadata request failed with status ${response.status}`);
	}

	return await response.json<FileMetadata>();
}

async function getR2Object(
	env: Env,
	key: string,
	rangeHeader: string | null,
): Promise<R2ObjectBody | null> {
	if (!rangeHeader) {
		return await env.NOTES_BUCKET.get(key);
	}

	const range = parseRangeHeader(rangeHeader);

	if (!range) {
		return await env.NOTES_BUCKET.get(key);
	}

	return await env.NOTES_BUCKET.get(key, {
		range,
	});
}

function parseRangeHeader(rangeHeader: string): R2Range | null {
	const match = rangeHeader.match(/^bytes=(\d*)-(\d*)$/);

	if (!match) {
		return null;
	}

	const startText = match[1];
	const endText = match[2];

	if (startText === "" && endText === "") {
		return null;
	}

	if (startText !== "" && endText !== "") {
		const start = Number(startText);
		const end = Number(endText);

		if (!Number.isSafeInteger(start) || !Number.isSafeInteger(end) || end < start) {
			return null;
		}

		return {
			offset: start,
			length: end - start + 1,
		};
	}

	if (startText !== "") {
		const start = Number(startText);

		if (!Number.isSafeInteger(start)) {
			return null;
		}

		return {
			offset: start,
		};
	}

	const suffix = Number(endText);

	if (!Number.isSafeInteger(suffix) || suffix <= 0) {
		return null;
	}

	return {
		suffix,
	};
}

function buildFileHeaders(metadata: FileMetadata, object: R2ObjectBody): Headers {
	const headers = new Headers();

	headers.set("Content-Type", metadata.file_content_type);
	headers.set("Accept-Ranges", "bytes");
	headers.set("Cache-Control", "public, max-age=3600");
	headers.set("ETag", object.etag);

	const fileName = safeDownloadFileName(metadata.original_file_name || `${metadata.title}.pdf`);
	headers.set("Content-Disposition", `inline; filename="${fileName}"`);

	if (object.range) {
		const range = object.range;

		if ("offset" in range && "length" in range && range.offset !== undefined && range.length !== undefined) {
			const start = range.offset;
			const end = range.offset + range.length - 1;
			headers.set("Content-Range", `bytes ${start}-${end}/${metadata.file_size_bytes}`);
			headers.set("Content-Length", String(range.length));
		}

		else if (object.size) {
			headers.set("Content-Length", String(object.size));
		}
	} else {
		headers.set("Content-Length", String(metadata.file_size_bytes));
	}

	return headers;
}

function safeDownloadFileName(fileName: string): string {
	return fileName
		.replace(/["\\]/g, "")
		.replace(/[^\w.\- ]+/g, "_")
		.trim()
		.slice(0, 120) || "notes.pdf";
}

function jsonError(
	status: number,
	code: string,
	message: string,
	extraHeaders?: HeadersInit,
): Response {
	return Response.json(
		{
			error: {
				code,
				message,
			},
		},
		{
			status,
			headers: extraHeaders,
		},
	);
}