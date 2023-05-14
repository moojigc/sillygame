<script lang="ts">
	import { onMount } from 'svelte';
	import { Coords, Pointer } from '../game/pointer';
	import {
		Events,
		type EventTypes,
		Keys,
		type Message
	} from '../game/constants';
	import { Messenger } from '../socket/messenger';
	import logger from '../logger/logger';
	import { writable, type Writable } from 'svelte/store';
	import Header from '../components/Header.svelte';

	let canvas: HTMLCanvasElement;

	const canvasSize: Writable<[string, string]> = writable(['720px', '900px']);
	const mouseDown: Writable<boolean> = writable(false);
	const debugMode: Writable<string> = writable('');
	const players: Writable<Record<string, any>> = writable({});
	const moves: Writable<Message[]> = writable([]);

	const keyDownMapping = {
		[Keys.Up]: (pt: Pointer) => ({ y: -pt.speed }),
		[Keys.Down]: (pt: Pointer) => ({ y: pt.speed }),
		[Keys.Left]: (pt: Pointer) => ({ x: -pt.speed }),
		[Keys.Right]: (pt: Pointer) => ({ x: pt.speed })
	};

	function getMousePos(
		canvas: HTMLCanvasElement,
		evt: { clientX: number; clientY: number }
	) {
		const rect = canvas.getBoundingClientRect(), // abs. size of element
			scaleX = canvas.width / rect.width, // relationship bitmap vs. element for x
			scaleY = canvas.height / rect.height; // relationship bitmap vs. element for y

		return {
			x: (evt.clientX - rect.left) * scaleX, // scale mouse coordinates after they have
			y: (evt.clientY - rect.top) * scaleY // been adjusted to be relative to element
		};
	}

	function MoveListener(
		pointer: Pointer,
		messenger: Messenger<EventTypes, Message>
	) {
		return (ev: { key: string }) => {
			if (ev.key in keyDownMapping) {
				const fn = keyDownMapping[ev.key];
				const { x, y } = pointer.getAbs(fn(pointer));
				messenger.send({
					event: Events.M,
					ts: Date.now(),
					mouseDown: true,
					coords: [x, y]
				});
			}
		};
	}

	function MouseListener(
		pointer: Pointer,
		messenger: Messenger<EventTypes, Message>
	) {
		return (ev: {
			clientX: number;
			clientY: number;
			preventDefault?: () => void;
		}) => {
			ev.preventDefault?.();
			const { x, y } = getMousePos(canvas, ev);

			const message: Message = {
				event: Events.M,
				coords: [x, y],
				ts: Date.now(),
				mouseDown: $mouseDown
			};
			messenger.send(message);

			pointer.moveAbs(Coords.fromVector(message.coords), $mouseDown);
		};
	}

	function getUser(id?: string, messenger?: Messenger<EventTypes, Message>) {
		if (id == messenger?.userId) {
			return `You (${id})`;
		}
		return `User ${id}`;
	}

	let messenger: Messenger<EventTypes, Message>;

	onMount(() => {
		const isMobile = window.matchMedia('(max-width: 900px)');

		isMobile.addEventListener('change', (evt) => {
			if (evt.matches) {
				$canvasSize = [
					`${window.innerWidth}px`,
					`${window.innerHeight - 200}px`
				];
			} else {
				$canvasSize = [`700px`, `${window.innerHeight - 200}px`];
			}
		});

		const ctx = canvas.getContext('2d')!;
		const pointer = new Pointer(ctx);

		messenger = new Messenger<EventTypes, Message>(
			import.meta.env.VITE_WEBSOCKET_URL
		);

		fetch(import.meta.env.VITE_WEB_URL + '/subscribers')
			.then(async (res) => {
				const subscribers: { id: string }[] = (await res.json())[
					'data'
				];

				$players = subscribers.reduce(
					(pv, s) => ({
						...pv,
						[s.id]: { id: s.id }
					}),
					{}
				);
			})
			.catch(console.error);

		document.addEventListener('keydown', MoveListener(pointer, messenger));
		document.addEventListener('mouseup', (ev) => void ($mouseDown = false));
		document.addEventListener(
			'mousedown',
			(ev) => void ($mouseDown = true)
		);

		document.addEventListener(
			'touchend',
			(ev) => void ($mouseDown = false)
		);
		document.addEventListener(
			'touchstart',
			(ev) => void ($mouseDown = true)
		);

		document.addEventListener('keypress', (evt) => {
			if ($debugMode.length > 5) {
				$debugMode = '';
			}
			$debugMode += evt.key;
		});

		canvas.addEventListener('mousemove', MouseListener(pointer, messenger));

		const listen = MouseListener(pointer, messenger);
		canvas.addEventListener('touchmove', (ev) => {
			$mouseDown = true;
			for (const t of ev.changedTouches) {
				listen(t);
			}
			$mouseDown = false;
		});

		messenger.onMessage(Events.HANDSHAKE, (data: Message<'HANDSHAKE'>) => {
			logger.debug('+page.svelte callback HANDSHAKE');
			messenger.register(data.userId!);
			$players = {
				...$players,
				[data.userId!]: { id: data.userId! }
			};
		});

		messenger.onMessage(Events.M, (data) => {
			logger.debug('+page.svelte callback', data);
			if (data.userId !== messenger.userId) {
				pointer.moveAbs(Coords.fromVector(data.coords), data.mouseDown);
			}
		});

		messenger.onMessage(Events.PLAYER_JOINED, (data) => {
			logger.debug('+page.svelte callback', data);

			$players = {
				...$players,
				[data.userId!]: { id: data.userId! }
			};
		});

		messenger.onMessage(Events.PLAYER_LEFT, (data) => {
			logger.debug('+page.svelte callback', data);

			delete $players[data.userId!];

			$players = { ...$players };
		});

		return () => {
			messenger.close();
		};
	});
</script>

<Header />
<div id="container" class="pure-g">
	<div>
		<h1>Welcome to a silly game.</h1>
		<canvas
			bind:this={canvas}
			id="canvas"
			width={$canvasSize[0]}
			height={$canvasSize[1]}
		/>
		<div id="players">
			{Object.keys($players).length} players joined.
		</div>
		{#if $debugMode.toLowerCase() == 'debug'}
			<div id="log">
				<h2>Move Log</h2>
				<ul>
					{#each $moves as move}
						<li>
							{#if move.event == Events.M}
								<p>
									<code>
										{getUser(move.userId, messenger)}
									</code>
									moved to {move.coords[0]}, {move.coords[1]}
								</p>
							{:else}
								<p>
									<code>
										{getUser(move.userId, messenger)}
									</code>
									joined!
								</p>
							{/if}
						</li>
					{/each}
				</ul>
			</div>
		{/if}
	</div>
</div>

<style>
	h1 {
		text-align: center;
	}
	canvas {
		border: 1px black solid;
		touch-action: none;
		margin-left: auto;
		margin-right: auto;
	}
	#container {
		display: flex;
		margin-left: auto;
		margin-right: auto;
		justify-content: center;
	}
	#players {
		padding: 1rem;
		text-align: center;
	}
	#log {
		text-align: center;
		max-width: 600px;
		max-height: 400px;
		overflow-y: scroll;
	}
	#log h2 {
		text-align: center;
	}
</style>
