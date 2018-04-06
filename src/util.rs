#[allow(dead_code)]
pub fn array_as_u32(array: &[u8; 4]) -> u32 {
    u32::from(array[0]) << 24 | u32::from(array[1]) << 16
        | u32::from(array[2]) << 8 | u32::from(array[3])
}

#[allow(dead_code)]
pub fn vec_as_u32(array: &[u8]) -> u32 {
    u32::from(array[0]) << 24 | u32::from(array[1]) << 16
        | u32::from(array[2]) << 8 | u32::from(array[3])
}

#[allow(dead_code)]
pub fn u32_as_vec(n: u32) -> Vec<u8> {
    let mut arr = Vec::with_capacity(4);
    arr[0] = (n >> 24) as u8;
    arr[1] = (n >> 16) as u8;
    arr[2] = (n >> 8) as u8;
    arr[3] = n as u8;
    arr
}

#[allow(dead_code)]
pub fn u32_into_vec(dest: &mut Vec<u8>, n: u32) {
    dest[0] = (n >> 24) as u8;
    dest[1] = (n >> 16) as u8;
    dest[2] = (n >> 8) as u8;
    dest[3] = n as u8;
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

pub trait VecExt {
    fn chew(&mut self, n: usize) -> Self;
}

impl<T> VecExt for Vec<T>
where
    T: Clone,
{
    fn chew(&mut self, n: usize) -> Self { self.drain(.. n).collect() }
}
