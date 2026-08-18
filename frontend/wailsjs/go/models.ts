export namespace dictionary {
	
	export class Entry {
	    Headword: string;
	    HTML: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Headword = source["Headword"];
	        this.HTML = source["HTML"];
	    }
	}

}

