export namespace ble {
	
	export class DeviceInfo {
	    name: string;
	    address: string;
	    rssi: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.address = source["address"];
	        this.rssi = source["rssi"];
	    }
	}

}

export namespace config {
	
	export class GlobalConfig {
	    default_device: string;
	    max_hr: number;
	    language: string;
	    auto_start: boolean;
	    check_update: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GlobalConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.default_device = source["default_device"];
	        this.max_hr = source["max_hr"];
	        this.language = source["language"];
	        this.auto_start = source["auto_start"];
	        this.check_update = source["check_update"];
	    }
	}
	export class WindowConfig {
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	    always_on_top: boolean;
	    opacity: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.always_on_top = source["always_on_top"];
	        this.opacity = source["opacity"];
	    }
	}
	export class Scene {
	    name: string;
	    window: WindowConfig;
	    style: string;
	    font_size: number;
	
	    static createFrom(source: any = {}) {
	        return new Scene(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.window = this.convertValues(source["window"], WindowConfig);
	        this.style = source["style"];
	        this.font_size = source["font_size"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace stats {
	
	export class SessionSummary {
	    id: string;
	    // Go type: time
	    start: any;
	    // Go type: time
	    end?: any;
	    duration: string;
	    avg_hr: number;
	    max_hr: number;
	    min_hr: number;
	
	    static createFrom(source: any = {}) {
	        return new SessionSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.start = this.convertValues(source["start"], null);
	        this.end = this.convertValues(source["end"], null);
	        this.duration = source["duration"];
	        this.avg_hr = source["avg_hr"];
	        this.max_hr = source["max_hr"];
	        this.min_hr = source["min_hr"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace style {
	
	export class Definition {
	    name: string;
	    version: number;
	    component: string;
	    author?: string;
	    default: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Definition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.component = source["component"];
	        this.author = source["author"];
	        this.default = source["default"];
	    }
	}

}

