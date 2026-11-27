Deno.test("the package loads over HTTP and renders", async () => {
	const types = { js: "text/javascript", wasm: "application/wasm" };
	const server = Deno.serve({ port: 0, onListen() {} }, async (request) => {
		const name = new URL(request.url).pathname.slice(1);
		const file = await Deno.open(new URL(name, import.meta.url));
		return new Response(file.readable, { headers: { "content-type": types[name.split(".").pop()] } });
	});
	const { load } = await import(`http://localhost:${server.addr.port}/index.js`);
	const fejkdata = await load();
	const person = fejkdata.render("{/sv_SE.person}", { seed: 1, today: "2026-01-15" });
	await server.shutdown();
	if (person !== "Felix Östberg") {
		throw new Error(`seed 1 rendered ${person}, want Felix Östberg`);
	}
});

Deno.test("the package loads from disk and renders", async () => {
	const { load } = await import("./index.js");
	const person = (await load()).render("{/sv_SE.person}", { seed: 1, today: "2026-01-15" });
	if (person !== "Felix Östberg") {
		throw new Error(`rendered ${person}`);
	}
});
