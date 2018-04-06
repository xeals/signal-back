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
use util::{self, VecExt};

pub struct BackupFile {
    pub file: Vec<u8>,
    pub cipher_key: Vec<u8>,
    pub mac_key: Vec<u8>,
    pub mac: Hmac<Sha256>,
    pub counter: u32,
    pub iv: Vec<u8>,
}

/// A backup is composed of the following:
///
/// + 4 bytes denoting the file header length (as a `u32`)
/// + X bytes, for the header
///   + The header contains the IV and salt for the encrypted backup
/// + X frames, composed of the following:
///   + 4 bytes denoting the frame length
///   + X bytes for the encoded frame itself
///   + 10 bytes for the original MAC of the encoded frame
impl BackupFile {
    pub fn new(p: &str, password: &str) -> io::Result<Self> {
        let mut f: Vec<u8> = ::std::fs::read(p)?;
        println!("file length: {}", f.len());
        println!("start of file: {:?}", &f[.. 8]);

        let header_length_bytes = f.chew(4);
        let header_length = util::vec_as_u32(&header_length_bytes);
        println!("header length: {}", header_length);

        let header_frame = f.chew(header_length as usize);
        println!("start of header: {:?}", &header_frame[.. 8]);

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

                println!("backup key: {:?}", key);
                println!("cipher key: {:?}", cipher_key);
                println!("mac key: {:?}", mac_key);

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

    pub fn next_frame(&mut self) -> io::Result<BackupFrame> {
        println!("start of frame: {:?}", &self.file[.. 8]);

        let length = self.file.chew(4);
        let frame_length = util::vec_as_u32(&length);
        println!("frame length: {}", frame_length);

        let mut frame = self.file.chew(frame_length as usize);
        println!("start of frame: {:?}", &frame[.. 8]);

        let _len = frame.len();
        let their_mac = frame.split_off(_len - 10);
        println!("remaining frame: {:?}", frame);

        self.mac.reset();
        self.mac.input(&frame);
        let our_mac = self.mac.result();

        if MacResult::new(&their_mac) == our_mac {
            Err(Error::new("Bad MAC"))?
        }

        let c = self.counter;
        util::u32_into_vec(&mut self.iv, c);
        self.counter = self.counter + 1;

        // assert_eq!(self.iv, [217, 193, 26, 204, 189, 32, 214, 84, 232, 116, 142, 68, 245, 144, 30, 31]);

        println!("new iv: {:?}", self.iv);

        let mut cipher = aes::ctr(
            aes::KeySize::KeySize128,
            &self.cipher_key,
            &self.iv,
        );

        // let aes_dec = ::crypto::aessafe::AesSafe128EncryptorX8::new(&self.cipher_key);
        // let mut cipher = ::crypto::blockmodes::CtrModeX8::new(aes_dec, &self.iv);

        // let mut cipher = ::crypto::blockmodes::CtrMode::new(
        //     ::crypto::aesni::AesNiEncryptor::new(
        //         aes::KeySize::KeySize128,
        //         &self.cipher_key
        //     ),
        //     self.iv.to_vec()
        // );

        let mut output = util::zeroed(frame.len());
        (*cipher).process(&frame, &mut output);
        // use crypto::symmetriccipher::Decryptor;
        // match cipher.decrypt(
        //     &mut ::crypto::buffer::RefReadBuffer::new(&frame),
        //     &mut ::crypto::buffer::RefWriteBuffer::new(&mut output),
        //     false
        // ) {
        //     Ok(_) => (),
        //     Err(_) => Err(Error::new("no decrypt for you"))?,
        // }
        // output = vec![42, 2, 8, 4];

        println!("decrypted: {:?}", output);

        BackupFrame::decode(output)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))
    }
}

fn backup_key(password: &str, salt: &[u8]) -> io::Result<Vec<u8>> {
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
    // Using a 30-byte input padded to 64 bytes is completely different to using
    // just the 30 bytes, so that's all we use for the first round.
    digest.input(&input);
    digest.input(&input);
    digest.result(&mut hash);
    digest.reset();

    println!("hash after round 1: {:?}", hash);

    for _ in 1 .. 250_000 {
        digest.input(&hash);
        digest.input(&input);
        digest.result(&mut hash);
        digest.reset();
    }

    hash.truncate(32);
    Ok(hash)
}
