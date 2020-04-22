extern crate hmac;
extern crate sha2;

use sha2::Sha256;
use hmac::{Hmac, Mac};

#[macro_use]
extern crate lazy_static;

use std::time::Instant;

// Create alias for HMAC-SHA256
type HmacSha256 = Hmac<Sha256>;

// https://lib.rs/crates/actix-web
use actix_web::{get, web, App, HttpServer, Responder};

#[derive(Debug)]
struct Global {
    work: u32,
}

impl Global {
    fn from_cli_args() -> Self {
    	// parse command line arg: usage:  `co-http work=N`
    	let arg1 = std::env::args().nth(1).expect("no work defined");
    	assert!(arg1.starts_with("work="));
    	let work = arg1[5..].parse::<u32>().unwrap();

        // store parsed args 
        Self { work }
    }
}

lazy_static! {
    static ref GLOBAL: Global = Global::from_cli_args();
}


#[get("/")]
async fn index(_info: web::Path<()>) -> impl Responder {
	let now = Instant::now();
	busy_calc(GLOBAL.work);
	let duration = now.elapsed();
	let usec = duration.as_secs() * 1_000_000 + duration.subsec_micros() as u64;    

	format!("busy for {} usec", usec)
}

#[actix_rt::main]
async fn main() -> std::io::Result<()> {
	// run web server
	let addr = "127.0.0.1:8080";
	println!("Listening at {}, will do {:?}", &addr, *GLOBAL);
    HttpServer::new(|| App::new().service(index))
        .bind(&addr)?
        .run()
        .await
}

/*fn main() {
	// parse command line arg: usage:  `co-http work=N`
	let arg1 = std::env::args().nth(1).expect("no work defined");
	assert!(arg1.starts_with("work="));
	let work = arg1[5..].parse::<u32>().unwrap();

	let now = Instant::now();
	busy_calc(work);
	let duration = now.elapsed();
	let usec = duration.as_secs() * 1_000_000 + duration.subsec_micros() as u64;
	println!("busy for {} us", usec);
}
*/

fn busy_calc(n: u32) {
	let msg = [0u8; 32];
	for _i in 0..n {
		// see https://docs.rs/hmac/0.7.1/hmac/
		let mut mac = HmacSha256::new_varkey(b"my secret and secure key")
		    .expect("HMAC can take key of any size");
		mac.input(&msg);
		mac.input(&msg);
		let result = mac.result();
		let _code_bytes = result.code();
	}
}

#[test]
fn test_calc() {
	let t1 = Instant::now();
	busy_calc(1000);
	let d1 = t1.elapsed();

	let t2 = Instant::now();
	busy_calc(2000);
	let d2 = t2.elapsed();

	assert!(d1 < d2)
}