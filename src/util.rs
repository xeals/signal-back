pub fn array_as_u32(array: &[u8; 4]) -> u32 {
    u32::from(array[0]) << 24 | u32::from(array[1]) << 16
        | u32::from(array[2]) << 8 | u32::from(array[3])
}

/// Pads a `Vec` with 0s to meet the desired length.
pub fn fill_to(v: Vec<u8>, length: usize) -> Vec<u8> {
    let mut vm = v;
    let len = vm.len();
    vm.extend(zeroed(length - len));
    vm
}

/// Returns a zeroed Vec to the specified length.
pub fn zeroed(len: usize) -> Vec<u8> {
    ::std::iter::repeat(0).take(len).collect()
}
