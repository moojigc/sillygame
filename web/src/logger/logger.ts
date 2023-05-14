enum Level {
	error,
	info,
	debug
}

export class Logger {
	constructor(
		public level: Level,
		private _logFn = (messages: any[]) => {
			const [meta, msgs] = this.format(messages);
			if (this.level == Level.error) {
				console.error(meta, ...msgs);
			} else {
				console.log(meta, ...msgs);
			}
		}
	) {}

	debug(...messages: any[]) {
		if (this.level < Level.debug) {
			return;
		}

		this._logFn(messages);
	}

	info(...messages: any[]) {
		if (this.level < Level.info) {
			return;
		}

		this._logFn(messages);
	}

	error(...messages: any[]) {
		if (this.level < Level.error) {
			return;
		}

		this._logFn(messages);
	}

	format(messages: any[]): [string, any[]] {
		return [
			`[${Level[this.level]} ${new Date().toLocaleTimeString()}]`,
			messages
		];
	}
}

const logger = new Logger(import.meta.env.VITE_LOG_LEVEL || Level.debug);
export default logger;
