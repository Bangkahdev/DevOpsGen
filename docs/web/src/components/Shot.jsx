export default function Shot({ rel, name, alt, caption, feature = '' }) {
  return (
    <figure className={`shot ${feature}`}>
      <a href={`${rel}img/${name}`} target="_blank" rel="noopener noreferrer">
        <img src={`${rel}img/${name}`} alt={alt} loading="lazy" decoding="async" />
      </a>
      {caption && <figcaption>{caption}</figcaption>}
    </figure>
  );
}
