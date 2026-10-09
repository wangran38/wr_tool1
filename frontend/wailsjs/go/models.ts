export namespace cleaner {
	
	export class Options {
	    ignoreWhitespace: boolean;
	    ignoreCase: boolean;
	    ignorePunctuation: boolean;
	    ignoreNumbering: boolean;
	    ignoreLineBreaks: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ignoreWhitespace = source["ignoreWhitespace"];
	        this.ignoreCase = source["ignoreCase"];
	        this.ignorePunctuation = source["ignorePunctuation"];
	        this.ignoreNumbering = source["ignoreNumbering"];
	        this.ignoreLineBreaks = source["ignoreLineBreaks"];
	    }
	}

}

export namespace differ {
	
	export class CellChange {
	    col: number;
	    kind: string;
	    old: string;
	    new: string;
	
	    static createFrom(source: any = {}) {
	        return new CellChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.col = source["col"];
	        this.kind = source["kind"];
	        this.old = source["old"];
	        this.new = source["new"];
	    }
	}
	export class Segment {
	    kind: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new Segment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.text = source["text"];
	    }
	}
	export class Pair {
	    kind: string;
	    old: string;
	    new: string;
	    oldSegs?: Segment[];
	    newSegs?: Segment[];
	
	    static createFrom(source: any = {}) {
	        return new Pair(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.old = source["old"];
	        this.new = source["new"];
	        this.oldSegs = this.convertValues(source["oldSegs"], Segment);
	        this.newSegs = this.convertValues(source["newSegs"], Segment);
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
	export class Stats {
	    equal: number;
	    insert: number;
	    delete: number;
	    replace: number;
	    similarity: number;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.equal = source["equal"];
	        this.insert = source["insert"];
	        this.delete = source["delete"];
	        this.replace = source["replace"];
	        this.similarity = source["similarity"];
	    }
	}
	export class Result {
	    pairs: Pair[];
	    stats: Stats;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pairs = this.convertValues(source["pairs"], Pair);
	        this.stats = this.convertValues(source["stats"], Stats);
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
	export class RowChange {
	    kind: string;
	    old: string[];
	    new: string[];
	    cells: CellChange[];
	
	    static createFrom(source: any = {}) {
	        return new RowChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.old = source["old"];
	        this.new = source["new"];
	        this.cells = this.convertValues(source["cells"], CellChange);
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
	
	
	export class TableDiff {
	    index: number;
	    kind: string;
	    rowCount: number;
	    rows: RowChange[];
	    stats: Stats;
	
	    static createFrom(source: any = {}) {
	        return new TableDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.kind = source["kind"];
	        this.rowCount = source["rowCount"];
	        this.rows = this.convertValues(source["rows"], RowChange);
	        this.stats = this.convertValues(source["stats"], Stats);
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

export namespace main {
	
	export class ComparisonResult {
	    oldFile: string;
	    newFile: string;
	    text: differ.Result;
	    tables: differ.TableDiff[];
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ComparisonResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldFile = source["oldFile"];
	        this.newFile = source["newFile"];
	        this.text = this.convertValues(source["text"], differ.Result);
	        this.tables = this.convertValues(source["tables"], differ.TableDiff);
	        this.truncated = source["truncated"];
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
	export class LicenseStatus {
	    machineCode: string;
	    activated: boolean;
	    trialDays: number;
	    expired: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LicenseStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.machineCode = source["machineCode"];
	        this.activated = source["activated"];
	        this.trialDays = source["trialDays"];
	        this.expired = source["expired"];
	    }
	}

}

