<script lang="ts">
	export let label: string;
	export let value = '';
	export let validate = (_value: string) => true;

	let focused = false;
	const focus = () => (focused = true);
	const blur = () => {
		focused = false;
		checkValue(value, true);
	};

	let hasError = false;
	const checkValue = (value: string, markNewError: boolean) => {
		if (validate(value)) {
			hasError = false;
			return;
		}
		if (markNewError) {
			hasError = true;
		}
	};

	$: checkValue(value, false);
</script>

<label class:hasError>
	<div class="label" class:expand={value === '' && !focused}>
		<div class="absolute left-0 top-0 w-max">{label}</div>
	</div>
	<input bind:value on:focus={focus} on:blur={blur} />
</label>

<style lang="postcss">
	label {
		@apply flex cursor-text flex-col text-gray-800;

		div.label {
			@apply relative h-4 w-0 text-xs transition-all;
			&.expand {
				@apply translate-x-1 translate-y-4 text-lg text-gray-500;
			}
		}
		input {
			@apply border-b border-gray-800 px-1 transition-colors;
			margin: 2px 2px 1px 2px;
		}
	}

	label:has(input:focus) {
		@apply text-blue-800;
		input {
			@apply border-2 border-blue-600 text-gray-800 outline-none;
			margin: 0;
		}
	}

	label.hasError {
		@apply text-red-800;

		div.label.expand {
			@apply text-red-800;
		}
		input {
			@apply border border-red-600 text-gray-800;
			margin: 1px;
		}
	}
</style>
