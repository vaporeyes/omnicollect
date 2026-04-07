export namespace main {
	
	export class DisplayHints {
	    label?: string;
	    placeholder?: string;
	    widget?: string;
	    group?: string;
	    order?: number;
	
	    static createFrom(source: any = {}) {
	        return new DisplayHints(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.placeholder = source["placeholder"];
	        this.widget = source["widget"];
	        this.group = source["group"];
	        this.order = source["order"];
	    }
	}
	export class AttributeSchema {
	    name: string;
	    type: string;
	    required?: boolean;
	    options?: string[];
	    display?: DisplayHints;
	
	    static createFrom(source: any = {}) {
	        return new AttributeSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.required = source["required"];
	        this.options = source["options"];
	        this.display = this.convertValues(source["display"], DisplayHints);
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
	export class BulkDeleteResult {
	    deleted: number;
	
	    static createFrom(source: any = {}) {
	        return new BulkDeleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deleted = source["deleted"];
	    }
	}
	export class BulkUpdateResult {
	    updated: number;
	
	    static createFrom(source: any = {}) {
	        return new BulkUpdateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updated = source["updated"];
	    }
	}
	
	export class Item {
	    id: string;
	    moduleId: string;
	    title: string;
	    purchasePrice?: number;
	    images: string[];
	    attributes: Record<string, any>;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.moduleId = source["moduleId"];
	        this.title = source["title"];
	        this.purchasePrice = source["purchasePrice"];
	        this.images = source["images"];
	        this.attributes = source["attributes"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ModuleSchema {
	    id: string;
	    displayName: string;
	    description?: string;
	    attributes: AttributeSchema[];
	
	    static createFrom(source: any = {}) {
	        return new ModuleSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	        this.description = source["description"];
	        this.attributes = this.convertValues(source["attributes"], AttributeSchema);
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
	export class ProcessImageResult {
	    filename: string;
	    originalPath: string;
	    thumbnailPath: string;
	    width: number;
	    height: number;
	    format: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessImageResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.originalPath = source["originalPath"];
	        this.thumbnailPath = source["thumbnailPath"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.format = source["format"];
	    }
	}

}

