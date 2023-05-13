<script lang="ts">
	import { onMount } from 'svelte';
	import { Coords, Pointer } from '../game/pointer';
	import { Keys } from '../game/keys';
	import { Messenger } from '../socket/messenger';
	import logger from '../logger/logger';
	import { writable, type Writable } from 'svelte/store';

	type EventTypes = 'M' | 'HANDSHAKE' | 'REGISTRATION' | 'PLAYER_LEFT';
	type MessageOrigin = 'S' | 'C';
	type Message<T extends EventTypes = EventTypes> = {
		source?: MessageOrigin;
		userId?: T extends 'HANDSHAKE' ? string : undefined;
		event: EventTypes;
		ts: number;
		coords: [number, number];
	};

	let canvas: HTMLCanvasElement;

	const players: Writable<Record<string, any>> = writable({});
	const moves: Writable<Message[]> = writable([]);

	const keyDownMapping = {
		[Keys.Up]: (pt: Pointer) => ({ y: -pt.speed }),
		[Keys.Down]: (pt: Pointer) => ({ y: pt.speed }),
		[Keys.Left]: (pt: Pointer) => ({ x: -pt.speed }),
		[Keys.Right]: (pt: Pointer) => ({ x: pt.speed })
	};

	function MoveListener(
		pointer: Pointer,
		messenger: Messenger<EventTypes, Message>
	) {
		return (ev: { key: string }) => {
			if (ev.key in keyDownMapping) {
				const fn = keyDownMapping[ev.key];
				pointer.moveRel(fn(pointer));
				messenger.send({
					event: 'M',
					ts: Date.now(),
					coords: [pointer.x, pointer.y]
				});
			}
		};
	}

	function getUser(id?: string, messenger?: Messenger<EventTypes, Message>) {
		if (id == messenger?.clientId) {
			return `You (${id})`;
		}
		return `User ${id}`;
	}

	let messenger: Messenger<EventTypes, Message>;

	onMount(() => {
		const ctx = canvas.getContext('2d')!;
		const pointer = new Pointer(ctx);
		messenger = new Messenger<EventTypes, Message>(
			import.meta.env.VITE_WEBSOCKET_URL
		);

		fetch(import.meta.env.VITE_WEB_URL + '/subscribers')
			.then(async (res) => {
				const subscribers: { id: string }[] = await res.json();

				$players = subscribers.reduce(
					(pv, s) => ({
						...pv,
						[s.id]: { id: s.id }
					}),
					{}
				);
			})
			.catch(console.error);

		const moveListener = MoveListener(pointer, messenger);
		// const [shiftDownListener, shiftUpListener] = ShiftListeners(pointer);

		document.addEventListener('keydown', moveListener);

		messenger.onMessage('HANDSHAKE', (data: Message<'HANDSHAKE'>) => {
			logger.debug('+page.svelte callback HANDSHAKE');
			messenger.register(data.userId!);
			$players = {
				...$players,
				[data.userId!]: { id: data.userId! }
			};
		});

		messenger.onMessage('M', (data) => {
			if (data.userId !== messenger.clientId) {
				pointer.moveAbs(Coords.fromVector(data.coords));
			}
			logger.debug('+page.svelte callback', data);
			// $moves = [data, ...$moves];
		});

		messenger.onMessage('REGISTRATION', (data) => {
			logger.debug('+page.svelte callback', data);

			$players = {
				...$players,
				[data.userId!]: { id: data.userId! }
			};
		});

		messenger.onMessage('PLAYER_LEFT', (data) => {
			logger.debug('+page.svelte callback', data);

			delete $players[data.userId!];

			$players = { ...$players };
		});

		return () => {
			messenger.close();
			document.removeEventListener('keydown', moveListener);
		};
	});
</script>

<div id="container" class="pure-g">
	<div>
		<h1 style="text-align: center">Welcome to a silly game.</h1>
		<h2>{Object.keys($players).length}</h2>
		<canvas bind:this={canvas} id="canvas" width="720" height="480" />
		<div id="log">
			<h2>Move Log</h2>
			<ul>
				{#each $moves as move}
					<li>
						{#if move.event == 'M'}
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
	</div>
</div>

<style>
	canvas {
		border: 1px black solid;
	}
	#container {
		display: flex;
		margin-left: auto;
		margin-right: auto;
		justify-content: center;
	}
	#log {
		max-width: 600px;
		max-height: 400px;
		overflow-y: scroll;
	}
	#log h2 {
		text-align: center;
	}
</style>
