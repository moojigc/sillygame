export const Keys = {
	Up: 'w',
	Down: 's',
	Left: 'a',
	Right: 'd',
	Shift: 'Shift'
};

export const DebugModeKeyCombo = ['d', 'e', 'b', 'u', 'g'];

export type EventTypes = 'M' | 'HANDSHAKE' | 'PLAYER_JOINED' | 'PLAYER_LEFT';

export const Events = {
	M: 'M' as const,
	HANDSHAKE: 'HANDSHAKE' as const,
	PLAYER_JOINED: 'PLAYER_JOINED' as const,
	PLAYER_LEFT: 'PLAYER_LEFT' as const
};

export type MessageOrigin = 'S' | 'C';
export type Message<T extends EventTypes = EventTypes> = {
	source?: MessageOrigin;
	userId?: T extends 'HANDSHAKE' ? string : undefined;
	event: EventTypes;
	ts: number;
	mouseDown?: boolean;
	coords: [number, number];
};
