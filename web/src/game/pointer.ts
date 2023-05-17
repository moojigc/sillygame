import logger from '../logger/logger';

const TOLERANCE_SCALE = 100;
class Moves {
	private _moves: Coords[] = [];
	add(move: Coords) {
		if (this._moves.length >= 16) {
			this._moves.shift();
		}
		this._moves.push(move);
	}

	hasEnoughData() {
		logger.debug(`moves: `, this._moves);
		return this._moves.length >= 8;
	}

	/**
	 * Get the greater average between x or y
	 */
	getGreatestAverage() {
		const diff: number[][] = [];
		for (let i = 1; i < this._moves.length; i++) {
			diff.push([
				Math.abs(this._moves[i].x - this._moves[i - 1].x),
				Math.abs(this._moves[i].y - this._moves[i - 1].y)
			]);
		}

		const [x, y] = diff.reduce(
			([px, py], [cx, cy]) => [px + cx, py + cy],
			[0, 0]
		);
		return [x / diff.length, y / diff.length];
	}
}

export class Coords {
	static fromMap({ x, y }: { x: number; y: number }) {
		return new Coords(x, y);
	}
	static fromVector([x, y]: [number, number]) {
		return new Coords(x, y);
	}
	constructor(public x: number, public y: number) {}
}
export class Pointer implements Coords {
	constructor(public renderContext: CanvasRenderingContext2D) {
		this.renderContext = renderContext;
		this.x = 0;
		this.y = 0;

		this._canvasCenter = {
			y: renderContext.canvas.height / 2,
			x: renderContext.canvas.width / 2
		};

		this.x = this._canvasCenter.x;
		this.y = this._canvasCenter.y;
		this.renderContext.moveTo(this._canvasCenter.x, this._canvasCenter.y);
		this.renderContext.beginPath();
	}

	x = 0;
	y = 0;
	speed = 2;
	lineWidth = 5;
	color = 'black';
	usingTouchscreen = false;

	private _canvasCenter: Coords;
	private _moves = new Moves();

	moveAbs({ x, y }: Coords, stroke = true) {
		const liftedUp = this._probablyLiftedFingerUp({ x, y });
		this.x = x;
		this.y = y;
		logger.info(this.x, this.y);

		if (stroke && !liftedUp) {
			this._drawLine();
		} else {
			this._move();
		}
		this._moves.add({ x: this.x, y: this.y });
	}

	getAbs({ x, y }: Partial<Coords>) {
		const absCoords = {
			x: (this.x = this.x + (x || 0)),
			y: (this.y = this.y + (y || 0))
		};

		return absCoords;
	}

	moveRel({ x, y }: Partial<Coords>, stroke = true) {
		const liftedUp = this._probablyLiftedFingerUp({
			x: this.x + (x || 0),
			y: this.y + (y || 0)
		});
		this.x = this.x + (x || 0);
		this.y = this.y + (y || 0);
		logger.info(this.x, this.y);

		if (stroke && !liftedUp) {
			this._drawLine();
		} else {
			this._move();
		}
		this._moves.add({ x: this.x, y: this.y });
	}

	private _probablyLiftedFingerUp({ x, y }: Coords) {
		if (!this._moves.hasEnoughData()) {
			return false;
		}
		const [distX, distY] = [Math.abs(x - this.x), Math.abs(y - this.y)];

		logger.debug('greaterDistance', [distX, distY]);

		return distX > TOLERANCE_SCALE || distY > TOLERANCE_SCALE;
	}

	private _move() {
		this.renderContext.moveTo(this.x, this.y);
	}

	private _drawLine() {
		this.renderContext.lineWidth = this.lineWidth;
		this.renderContext.strokeStyle = this.color;
		this.renderContext.lineTo(this.x, this.y);
		this.renderContext.stroke();
	}
}
