export namespace dictionary {
	
	export class Entry {
	    Dictionary: string;
	    Headword: string;
	    HTML: string;
	    CSS: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Dictionary = source["Dictionary"];
	        this.Headword = source["Headword"];
	        this.HTML = source["HTML"];
	        this.CSS = source["CSS"];
	    }
	}

}

export namespace main {
	
	export class PackageInfo {
	    Path: string;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new PackageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Name = source["Name"];
	    }
	}

}

