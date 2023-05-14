import logger from '../logger/logger';
import { v4 as uuidv4 } from 'uuid';

function getProtocol(): 'ws:' | 'wss:' {
	if (window.location.protocol == 'https:') {
		return 'wss:';
	}
	return 'ws:';
}

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export class MessengerError extends Error {}

export class Messenger<EventTypes extends string, Message> {
	constructor(public address: string, public protocols = ['silly-game']) {
		if (address.includes('ws')) {
			throw new MessengerError(
				'Address must not include ws/wss protocol'
			);
		}
		this.socket = this._getNewSocket();
	}

	userId: string | null = null;

	private _getNewSocket() {
		const ws = new WebSocket(getProtocol() + this.address, this.protocols);
		ws.onopen = function (ev) {
			logger.info('socket open', ev);
		};
		ws.onclose = function (ev) {
			logger.debug('socket closed', ev.code, ev.reason);
		};
		ws.onerror = function (ev) {
			logger.error('socket error', ev);
		};

		ws.onmessage = async (ev) => {
			let parsed;
			try {
				parsed = JSON.parse(ev.data);
			} catch (e) {
				throw new MessengerError(`Messenger JSON.parse err: ${e}`);
			}

			logger.info(`Messenger#onMessage(type: ${parsed.event}): `, parsed);
			const callback = this.callbacks.get(parsed.event);

			if (callback) {
				await callback(parsed);
			}
		};
		return ws;
	}

	socket!: WebSocket;
	closed = false;
	private callbacks: Map<
		EventTypes,
		(data: Message) => void | Promise<void>
	> = new Map();

	register(clientId: string) {
		logger.info('client id is ', clientId);
		this.userId = clientId;
	}

	async send(
		req: Message & {
			id?: string;
			ts?: number;
			userId?: string;
		},
		attempts = 0,
		sleepMs = 0
	): Promise<void> {
		logger.debug(
			`Messenger#send, attempt ${attempts}; waiting for ${
				sleepMs / 1000
			} seconds`
		);
		if (attempts > 3) {
			return;
		}

		if (sleepMs) {
			await sleep(sleepMs);
		}

		req['id'] = uuidv4();
		req['ts'] = Date.now();
		if (this.userId) {
			req['userId'] = this.userId!;
		}

		const msg = JSON.stringify(req);
		logger.debug(`Messenger#send`, msg);

		switch (this.socket.readyState) {
			case WebSocket.CONNECTING:
				return this.send(req, attempts + 1, sleepMs + 100 ** attempts);
			case WebSocket.CLOSED:
			case WebSocket.CLOSING:
				this.socket = this._getNewSocket();
				return this.send(req, attempts + 1, sleepMs + 100 ** attempts);
			default:
				this.socket.send(msg);
		}
	}

	onMessage(
		type: EventTypes,
		callback: (data: Message) => void | Promise<void>
	) {
		if (!this.callbacks.has(type)) {
			this.callbacks.set(type, callback);
		}
	}

	close(code?: number) {
		this.closed = true;
		this.socket.close(code);
	}
}
