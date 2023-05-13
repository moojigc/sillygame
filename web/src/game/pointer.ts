import logger from '../logger/logger';

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
	constructor(public ctx: CanvasRenderingContext2D) {
		this.ctx = ctx;
		this.x = 0;
		this.y = 0;

		this._canvasCenter = {
			y: ctx.canvas.height / 2,
			x: ctx.canvas.width / 2
		};

		this.x = this._canvasCenter.x;
		this.y = this._canvasCenter.y;
		this.ctx.moveTo(this._canvasCenter.x, this._canvasCenter.y);
		this.ctx.beginPath();
	}

	x = 0;
	y = 0;
	speed = 2;

	private _canvasCenter: Coords;

	moveAbs({ x, y }: Coords) {
		this.x = x;
		this.y = y;
		logger.info(this.x, this.y);

		this._drawLine();
	}

	moveRel({ x, y }: Partial<Coords>) {
		this.x = this.x + (x || 0);
		this.y = this.y + (y || 0);
		logger.info(this.x, this.y);

		this._drawLine();
	}

	private _drawLine(lineWidth = 5) {
		this.ctx.lineTo(this.x, this.y);
		this.ctx.lineWidth = lineWidth;
		this.ctx.stroke();
	}
}
