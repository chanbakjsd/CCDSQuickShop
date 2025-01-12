<script lang="ts">
	export let label: string;
	export let value = '';
	export let validate: (_value: string) => Promise<boolean> | boolean = () => true;
	export let invalid = false;

	let focused = false;
	const focus = () => (focused = true);
	const blur = () => {
		focused = false;
		checkValue(value, true);
	};

	let hasError = false;
	const checkValue = async (value: string, markNewError: boolean) => {
		if (await validate(value)) {
			hasError = false;
			return;
		}
		if (markNewError) {
			hasError = true;
		}
	};

	$: checkValue(value, false);
</script>

<label class="relative" class:hasError={hasError || invalid}>
	<div class="label" class:expand={value === '' && !focused}>
		<div class="absolute left-0 top-0 w-max">{label}</div>
	</div>
	<input bind:value on:focus={focus} on:blur={blur} />
	<div class="pointer-events-none ml-1 mr-2 mt-4 text-lg opacity-0" aria-hidden="true">{label}</div>
</label>

<style lang="postcss">
	label {
		@apply flex cursor-text flex-col text-gray-800;

		div.label {
			@apply z-10 w-0 text-xs transition-all;
			&.expand {
				@apply translate-x-1 translate-y-4 text-lg text-gray-500;
			}
		}
		input {
			@apply absolute border-b border-gray-800 bg-transparent px-1 transition-colors;
			top: calc(1rem + 2px);
			bottom: 1px;
			right: 2px;
			left: 2px;
		}
	}

	label:has(input:focus) {
		@apply text-blue-800;
		input {
			@apply border-2 border-blue-600 text-gray-800 outline-none;
			@apply bottom-0 left-0 right-0 top-4;
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
