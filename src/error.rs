use std::io;

pub type Result<T> = io::Result<T>;

pub struct Error(&'static str);

impl Error {
    pub fn new(s: &'static str) -> Error { Error(s) }
}

impl From<Error> for io::Error {
    fn from(e: Error) -> io::Error {
        io::Error::new(io::ErrorKind::InvalidData, e.0)
    }
}
