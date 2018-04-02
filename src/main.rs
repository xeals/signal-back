#![feature(fs_read_write_bytes)]

extern crate byteorder;
extern crate bytes;
extern crate crypto;
extern crate prost;
#[macro_use]
extern crate prost_derive;

mod decrypt;
mod signal;
mod hkdf3;
mod util;
mod error;

fn main() {
    let secret = ::std::fs::read_to_string("./signal.backup.password").unwrap();
    let mut file =
        decrypt::BackupFile::new("./signal.backup", &secret).unwrap();
    let bf = file.frame();

    println!("{:?}", bf);
}
