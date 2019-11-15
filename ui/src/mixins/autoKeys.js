export default {
	methods: {
		formatStringAsKey(str) {
		  return str
			.toLowerCase()
			.split(/\s+/)
			.join("-");
		}
	}
}
