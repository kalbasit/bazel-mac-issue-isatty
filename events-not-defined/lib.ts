import { EventEmitter } from "events";

const focusBlurEventEmitter = new EventEmitter();
focusBlurEventEmitter.emit('focus');
