use std::io::{self, Cursor, Read};

use byteorder::BigEndian;
use bytes::Buf;
use crypto::aes;
use crypto::digest::Digest;
use crypto::hmac::Hmac;
use crypto::mac::{Mac, MacResult};
use crypto::sha2::{Sha256, Sha512};
use prost::Message;

use error::Error;
use hkdf3;
use signal::BackupFrame;
use util;

pub struct BackupFile {
    pub file: Cursor<Vec<u8>>,
    pub cipher_key: Vec<u8>,
    pub mac_key: Vec<u8>,
    pub mac: Hmac<Sha256>,
    pub counter: u32,
    pub iv: Vec<u8>,
}

impl BackupFile {
    pub fn new(p: &str, password: &str) -> io::Result<Self> {
        let mut f = Cursor::new(::std::fs::read(p)?);

        let mut header_length_bytes = [0u8; 4];
        f.read_exact(&mut header_length_bytes)?;

        let header_length =
            Cursor::new(header_length_bytes).get_u32::<BigEndian>();
        let mut header_frame = Vec::with_capacity(header_length as usize);
        let mut header_f = Read::take(f.clone(), header_length.into());
        header_f.read_to_end(&mut header_frame)?;

        match BackupFrame::decode(header_frame) {
            Ok(frame) => {
                let header = frame.header.ok_or_else(|| {
                    Error::new("Backup file does not start with header!")
                })?;

                let iv = header
                    .iv
                    .ok_or_else(|| Error::new("No IV in header"))?;

                if iv.len() != 16 {
                    return Err(Error::new("Invalid IV length!"))?
                }

                let key =
                    backup_key(password, &header.salt.unwrap_or_default())?;
                let mut cipher_key =
                    hkdf3::derive_secrets(&key, b"Backup Export");
                let mac_key = cipher_key.split_off(32);

                assert_eq!(mac_key.len(), cipher_key.len());

                Ok(BackupFile {
                    file: f,
                    mac: Hmac::new(Sha256::new(), &mac_key),
                    cipher_key,
                    mac_key,
                    iv: iv.clone(),
                    counter: Cursor::new(iv).get_u32::<BigEndian>(),
                })
            }
            Err(e) => Err(io::Error::new(io::ErrorKind::InvalidData, e)),
        }
    }

    pub fn frame(&mut self) -> io::Result<BackupFrame> {
        let mut length = [0u8; 4];
        self.file.read_exact(&mut length)?;

        let frame_length = Cursor::new(length).get_u16::<BigEndian>();
        let mut header_f = Read::take(self.file.clone(), frame_length.into());

        let mut frame = Vec::with_capacity(frame_length as usize);
        header_f.read_to_end(&mut frame)?;

        // Here the Java version initialises the MAC decoder and compares it
        // using some inner working of how Java handles these things. Long story
        // short, Rust doesn't do that, so we can't verify anything in this
        // step.

        // let _len = frame.len();
        // let their_mac = frame.split_off(_len - 10);

        // self.mac.reset();
        // self.mac.input(&frame);
        // let our_mac = self.mac.result();

        // if MacResult::new(&their_mac) != our_mac {
        //     Err(Error::new("Bad MAC"))?
        // }

        let c = self.counter;
        util::u32_into_vec(&mut self.iv, c);

        let mut cipher = aes::ctr(
            aes::KeySize::KeySize128,
            &self.cipher_key,
            &self.iv,
        );

        let mut output = util::zeroed(frame.len());
        (*cipher).process(&frame, &mut output);

        BackupFrame::decode(output)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))
    }
}

fn backup_key(password: &str, salt: &[u8]) -> Result<Vec<u8>, io::Error> {
    let mut digest = Sha512::new();
    let input: Vec<u8> = password
        .trim()
        .replace(" ", "")
        .bytes()
        .collect();
    // 0-padded to 64 bytes
    // let mut hash: Vec<u8> = util::fill_to(input.clone(), 64);
    // let mut hash: Vec<u8> = input.clone();
    let mut hash = util::zeroed(64);

    if !salt.is_empty() {
        digest.input(salt);
    }

    // Do the first digest manually.
    // Reasoning is that the zeroed bytes may throw off the algorithm.
    digest.input(&input);
    digest.input(&input);
    digest.result(&mut hash);
    digest.reset();

    for _ in 1 .. 250_000 {
        digest.input(&hash);
        digest.input(&input);
        digest.result(&mut hash);
        digest.reset();
    }

    hash.truncate(32);
    Ok(hash)
}
